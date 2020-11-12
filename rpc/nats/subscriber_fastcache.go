package nats

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
)

// SubscriberFastCache The NatS Subscriber with Fast Cache Temporary Storage.
type SubscriberFastCache struct {
	*nats.Conn
	sub  *nats.Subscription
	Subj string
	Hand func([][]byte) error
	*fastcache.Cache
	CacheDir     string // sets cache persist to disk directory
	Index        uint64
	Count        uint64
	Since        *f.TimeStamp
	MsgLimit     int   // sets the limits for pending messages for this subscription.
	BytesLimit   int   // sets the limits for a message's bytes for this subscription.
	OnceAmount   int64 // sets amount allocated at one time
	OnceInterval time.Duration
	Running      bool
	async        bool
	err          error
	pool         *f.Pool
}

// NewSubscriberFastCache Create a subscriber with cache store for Client Connect.
func NewSubscriberFastCache(nc *nats.Conn, subject string, cacheDir ...string) *SubscriberFastCache {
	dir, since := ".nats", f.TimeFrom(time.Now(), true)

	if len(cacheDir) == 1 && cacheDir[0] != "" {
		if err := f.Mkdir(cacheDir[0]); err != nil {
			return nil
		}
		dir = filepath.Join(cacheDir[0], since.LocalTimeStampString(true))
	} else {
		if err := f.MkdirCurrent(dir); err != nil {
			return nil
		}
		dir = filepath.Join(f.CurrentDir(), dir, since.LocalTimeStampString(true))
	}

	client := fastcache.New(1073741824) // 1GB cache capacity

	sub := &SubscriberFastCache{
		Conn:         nc,
		Subj:         subject,
		Cache:        client, // fast cache
		CacheDir:     dir,
		Since:        since,
		MsgLimit:     100000000, // pending messages: 100 million
		BytesLimit:   1048576,   // a message's size: 1MB
		OnceAmount:   -1,
		OnceInterval: time.Second,
		async:        true,
	}

	sub.pool = f.NewPool(f.NumCPUx16, func() f.PoolWorker {
		return &CachePoolWorker{
			processor: sub.Process,
		}
	})

	return sub
}

// LimitMessage sets amount for pending messages for this subscription, and a message's bytes.
// Defaults amountPendingMessages: 100 million, anMessageBytes: 1MB
func (sub *SubscriberFastCache) LimitMessage(amountPendingMessages, anMessageBytes int) {
	sub.MsgLimit, sub.BytesLimit = amountPendingMessages, anMessageBytes
}

// LimitAmount sets amount allocated at one time, and the processing interval time.
// Defaults onceAmount: -1, onceInterval: time.Second
func (sub *SubscriberFastCache) LimitAmount(onceAmount int64, onceInterval time.Duration) {
	sub.OnceAmount, sub.OnceInterval = onceAmount, onceInterval
}

// Process messages for this subscription.
func (sub *SubscriberFastCache) Process(msg *CacheMsg) error {
	key, val := msg.Key, msg.Val
	sub.Cache.Set(f.BytesUint64(key), val)
	return nil
}

// Run runtime to end your application.
func (sub *SubscriberFastCache) Run(waitFunc ...func()) {
	sub.Running = true
	ctx, cancel := context.WithCancel(context.Background())

	// Handle panic.
	defer func() {
		err := recover()
		if err != nil {
			Log.Error().Msgf("[nats] stop receive new data with error.panic > %v", err)
		} else {
			Log.Warn().Msgf("[nats] stop receive new data > %d/%d < %d records not processed", sub.Index, sub.Count, sub.Count-sub.Index)
		}

		// Unsubscribe will remove interest in the given subject.
		_ = sub.sub.Unsubscribe()
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = sub.Conn.Drain()

		// Stop handle new data.
		cancel()
		// Stop pool processor.
		sub.pool.Close()
		// Save not processed data.
		sub.Save(sub.CacheDir)

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
		sub.Running = false
	}()

	// Async Subscriber.
	sub.sub, sub.err = sub.Conn.Subscribe(sub.Subj, func(msg *nats.Msg) {
		key, val := atomic.AddUint64(&sub.Count, 1), msg.Data
		//sub.pool.Process(&CacheMsg{Key: key, Val: val}) // It's slow
		sub.Cache.Set(f.BytesUint64(key), val)
	})
	// Set listening.
	SubscribeErrorHandle(sub.sub, sub.async, sub.err)
	if sub.err != nil {
		log.Fatal(sub.err)
	}

	// Set pending limits.
	SubscribeLimitHandle(sub.sub, sub.MsgLimit, sub.BytesLimit)

	// Flush connection to server, returns when all messages have been processed.
	FlushAndCheckLastError(sub.Conn)

	// init handle old data.
	go sub.init(ctx)

	// run handle new data.
	go sub.hand(ctx)

	if len(waitFunc) > 0 {
		Log.Info().Msg("[nats] start running and wait to auto exit")
		waitFunc[0]()
		return
	}

	Log.Info().Msg("[nats] start running and wait to manual exit")

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		Log.Warn().Msg("[nats] forced to shutdown.")
	})
}

// init handle old data.
func (sub *SubscriberFastCache) init(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	cacheDir := sub.CacheDir
	dir0, parentDir := filepath.Base(cacheDir), filepath.Dir(cacheDir)
	oldDirs, err := filepath.Glob(filepath.Join(parentDir, "*"))
	if err != nil || oldDirs == nil || len(oldDirs) == 0 {
		return
	}
	sort.Strings(oldDirs)

	var clearCache = func(cache *fastcache.Cache, index, count int64) {
		for i, c := uint64(index)+1, uint64(count); i <= c; i++ {
			k := f.BytesUint64(i)
			cache.Del(k)
		}
	}

	handRecords, oldRecords, onceRecords := 0, 0, atomic.LoadInt64(&sub.OnceAmount)
	var runHandle = func(dir string, dirname string) (ok bool) {
		oldFiles, err := filepath.Glob(filepath.Join(dir, "*"))
		if err != nil || oldFiles == nil || len(oldFiles) == 0 {
			ok = true
			return
		}
		sort.Strings(oldFiles)
		for _, oldFile := range oldFiles {
			_, jsonFile := filepath.Split(oldFile)
			if ok, _ := regexp.MatchString(`^\d+\.\d+\.\d+\.json$`, jsonFile); !ok {
				continue
			}

			fileDir := strings.Replace(jsonFile, ".json", "", 1)
			filePath := filepath.Join(dir, fileDir)
			if f.PathExists(filePath) == false {
				continue
			}

			cache, err := fastcache.LoadFromFile(filePath)
			s := strings.Split(fileDir, ".")
			if err != nil || len(s) != 3 {
				continue
			}

			since := s[0]
			index, _ := strconv.ParseInt(s[1], 10, 0)
			count, _ := strconv.ParseInt(s[2], 10, 0)
			indexZero, indexSize := uint64(index)+1, count-index
			if onceRecords > 0 {
				indexSize = onceRecords
			}

			var handData = make([][]byte, 0, indexSize)
			for i, c, dataIndex := indexZero, uint64(count), int64(0); i <= c; i++ {
				if key := f.BytesUint64(i); cache.Has(key) {
					val := cache.Get(nil, key)
					handData = append(handData, val)
					if dataIndex++; dataIndex == onceRecords || i == c {
						// bulk handle
						if err := sub.Hand(handData); err != nil {
							// rollback
							Log.Error().Msgf("[nats] init handle old data > %s > %s", dirname, err)
							if i > indexZero {
								clearCache(cache, int64(indexZero)-1, int64(i))
								dirname1, filename1 := sub.dirnames(since, i-1, c), sub.filenames(since, i-1, c)
								saveFastCache(cache, dir, dirname1, filename1)
								_ = os.Remove(oldFile)
								_ = os.RemoveAll(filePath)
							}
							return
						}
						handRecords += len(handData)
						// reset data
						dataIndex = 0
						handData = make([][]byte, 0, indexSize)
						time.Sleep(sub.OnceInterval)
					}
				}
			}

			_ = os.Remove(oldFile)
			_ = os.RemoveAll(filePath)
		}

		ok = true
		return
	}

	for _, oldDir := range oldDirs {
		if !f.IsDir(oldDir) {
			continue
		}

		dir1 := filepath.Base(oldDir)
		if ok, _ := regexp.MatchString(`^\d+$`, dir1); !ok {
			continue
		}
		if len(dir1) != 14 || dir1 == dir0 {
			continue
		}
		if datFiles, err := filepath.Glob(filepath.Join(oldDir, "*.json")); err != nil || datFiles == nil || len(datFiles) == 0 {
			continue
		}

		// reboot init handle old data
		for f.PathExists(oldDir) {
			if runHandle(oldDir, dir1) {
				if err1 := os.RemoveAll(oldDir); err1 != nil {
					Log.Error().Msgf("[nats] remove old data directory > %s > %s", dir1, err1)
				}
			}
			Log.Info().Msgf("[nats] init handle old data > %d records < %s", handRecords, dir1)
			oldRecords += handRecords
			handRecords = 0
		}
	}

	if oldRecords > 0 {
		Log.Info().Msgf("[nats] init handle old data > %d records", oldRecords)
	}

	if err := ctx.Err(); err != nil && err != context.Canceled {
		Log.Warn().Msgf("[nats] init handle old data err > %s", err)
	}
}

// run handle new data.
func (sub *SubscriberFastCache) hand(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	var (
		running  bool
		runCount uint64
		delIndex uint64
	)

	var runHandle = func() {
		count, index := atomic.LoadUint64(&sub.Count), atomic.LoadUint64(&sub.Index)
		if count <= runCount {
			// reset handle
			if 0 < count && 3 == time.Now().Hour() && index <= delIndex {
				sub.Count, runCount, sub.Index, delIndex = 0, 0, 0, 0
			}
			return
		}

		handRecords, onceRecords := int64(0), atomic.LoadInt64(&sub.OnceAmount)
		indexSize := int64(count - index)
		if onceRecords > 0 {
			indexSize = onceRecords
		}

		var handData = make([][]byte, 0, indexSize)
		for dataIndex := int64(0); dataIndex < indexSize && index < count; runCount++ {
			index++ // key equals index
			val := sub.Cache.Get(nil, f.BytesUint64(index))
			if val == nil {
				val = []byte{}
			}
			handData = append(handData, val)
			handRecords++
			if dataIndex++; dataIndex == indexSize || index == count {
				// bulk handle
				if err := sub.Hand(handData); err != nil {
					// rollback
					Log.Error().Msgf("[nats] run handle new data err > %s", err)
					handRecords -= dataIndex
					break
				}
				sub.Index += uint64(dataIndex)
				// reset data
				dataIndex = 0
				handData = make([][]byte, 0, indexSize)
			}
		}

		if handRecords == 0 {
			return
		}

		Log.Info().Msgf("[nats] run handle new data > %d/%d < %d records", sub.Index, sub.Count, handRecords)

		// delete old data
		go func(index, handRecords uint64) {
			for i, n := index-handRecords+1, index; i <= n; i++ {
				sub.Cache.Del(f.BytesUint64(i))
				atomic.AddUint64(&delIndex, 1)
			}
		}(sub.Index, uint64(handRecords))
	}

	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil && err != context.Canceled {
				Log.Warn().Msgf("[nats] done handle new data err > %s", err)
			}
			for running {
				time.Sleep(time.Millisecond)
			}
			runHandle()
			return
		case <-time.After(sub.OnceInterval):
			if running {
				continue
			}
			running = true
			Log.Debug().Msgf("[nats] run receive new data > %d/%d", sub.Index, sub.Count)
			runHandle()
			running = false
		}
	}
}

// Dirname gets the Cache Dirname.
func (sub *SubscriberFastCache) Dirname() string {
	return sub.dirname(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) dirname(since *f.TimeStamp, index, count uint64) string {
	return sub.dirnames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) dirnames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d", since, index, count)
}

// Filename gets the Cache Filename.
func (sub *SubscriberFastCache) Filename() string {
	return sub.filename(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) filename(since *f.TimeStamp, index, count uint64) string {
	return sub.filenames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) filenames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d.json", since, index, count)
}

// Save the Cache Data of some records not processed.
func (sub *SubscriberFastCache) Save(cacheDir string) {
	if sub.Count == 0 || sub.Hand == nil || sub.Index == sub.Count {
		return
	}

	saveFastCache(sub.Cache, cacheDir, sub.Dirname(), sub.Filename())
}

func saveFastCache(cache *fastcache.Cache, cacheDir, dirname, filename string) {
	fileStat := new(fastcache.Stats)
	cache.UpdateStats(fileStat)
	handData, err := f.EncodeJson(fileStat)
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats > %s", err)
	}

	filePath := filepath.Join(cacheDir, filename)
	err = ioutil.WriteFile(filePath, handData, 0644)
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats > %s", err)
	}

	dirPath := filepath.Join(cacheDir, dirname)
	if err = cache.SaveToFileConcurrent(dirPath, 0); err != nil {
		Log.Error().Msgf("[nats] save cache data > %s", err)
	} else {
		cache.Reset() // Reset removes all the items from the cache.
	}
}
