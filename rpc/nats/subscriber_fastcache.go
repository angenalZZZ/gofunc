package nats

import (
	"context"
	"fmt"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
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
)

const BulkSize = 2000

type SubscriberFastCache struct {
	*nats.Conn
	sub  *nats.Subscription
	Subj string
	Hand func([BulkSize][]byte) error
	*fastcache.Cache
	CacheDir   string // sets cache persist to disk directory
	Index      uint64
	Count      uint64
	Since      *f.TimeStamp
	MsgLimit   int // sets the limits for pending messages for this subscription.
	BytesLimit int // sets the limits for a message's bytes for this subscription.
	async      bool
	err        error
	pool       *f.Pool
}

// NewSubscriberFastCache Create a subscriber with cache store for Client Connect.
func NewSubscriberFastCache(nc *nats.Conn, subject string, cacheDir ...string) *SubscriberFastCache {
	sub := &SubscriberFastCache{
		Conn:       nc,
		Subj:       subject,
		Cache:      fastcache.New(1073741824), // 1GB cache capacity
		Since:      f.TimeFrom(time.Now(), true),
		MsgLimit:   100000000, // pending messages: 100 million
		BytesLimit: 10485760,  // a message's size: 10MB
		async:      true,
	}
	sub.pool = f.NewPool(f.NumCPUx16, func() f.PoolWorker {
		return &CachePoolWorker{
			processor: sub.Process,
		}
	})
	if len(cacheDir) == 1 && cacheDir[0] != "" {
		sub.CacheDir = cacheDir[0]
	} else {
		sub.CacheDir = data.CurrentDir
	}
	return sub
}

// Process messages for this subscription.
func (sub *SubscriberFastCache) Process(msg *CacheMsg) error {
	key, val := msg.Key, msg.Val
	sub.Cache.Set(f.BytesUint64(key), val)
	return nil
}

// Run runtime to end your application.
func (sub *SubscriberFastCache) Run(waitFunc ...func()) {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle panic
	defer func() {
		err := recover()
		if err != nil {
			Log.Error().Msgf("[nats] stop receive new data with error > %v", err)
		} else {
			Log.Info().Msgf("[nats] stop receive new data > %d/%d", sub.Index, sub.Count)
		}

		// Unsubscribe will remove interest in the given subject.
		_ = sub.sub.Unsubscribe()
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = sub.Conn.Drain()

		// Stop handle new data.
		cancel()
		// Stop pool processor.
		sub.pool.Close()
		// Save cache.
		sub.Save(sub.CacheDir)

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber
	sub.sub, sub.err = sub.Conn.Subscribe(sub.Subj, func(msg *nats.Msg) {
		key, val := atomic.AddUint64(&sub.Count, 1), msg.Data
		//sub.pool.Process(&CacheMsg{Key: key, Val: val})
		sub.Cache.Set(f.BytesUint64(key), val)
	})
	SubscribeErrorHandle(sub.sub, sub.async, sub.err)
	if sub.err != nil {
		log.Fatal(sub.err)
	}

	// Set pending limits
	SubscribeLimitHandle(sub.sub, sub.MsgLimit, sub.BytesLimit)

	// Flush connection to server, returns when all messages have been processed.
	FlushAndCheckLastError(sub.Conn)

	// init handle old data
	go sub.init(ctx)

	// run handle new data
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
		Log.Error().Msg("[nats] run forced termination")
	})
}

// init handle old data.
func (sub *SubscriberFastCache) init(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	oldFiles, _ := filepath.Glob(filepath.Join(sub.CacheDir, "*"))
	sort.Strings(oldFiles)

	cacheDir := sub.CacheDir
	var clearCache = func(cache *fastcache.Cache, index, count int64) {
		for i, c := uint64(index)+1, uint64(count); i <= c; i++ {
			k := f.BytesUint64(i)
			cache.Del(k)
		}
	}

	handRecords := 0
	for _, oldFile := range oldFiles {
		dir, jsonFile := filepath.Split(oldFile)
		if ok, _ := regexp.MatchString(`^\d+\.\d+\.\d+\.json$`, jsonFile); !ok {
			continue
		}

		dirname := strings.ReplaceAll(jsonFile, ".json", "")
		filePath := filepath.Join(dir, dirname)
		if f.PathExists(filePath) == false {
			continue
		}

		cache, err := fastcache.LoadFromFile(filePath)
		s := strings.Split(dirname, ".")
		if err != nil || len(s) != 3 {
			continue
		}

		since := s[0]
		index, _ := strconv.ParseInt(s[1], 10, 0)
		count, _ := strconv.ParseInt(s[2], 10, 0)
		indexZero := uint64(index) + 1

		var handData [BulkSize][]byte
		for i, c, dataIndex := indexZero, uint64(count), 0; i <= c; i++ {
			if key := f.BytesUint64(i); cache.Has(key) {
				handData[dataIndex] = cache.Get(nil, key)
				if dataIndex++; dataIndex == BulkSize || i == c {
					// bulk handle
					if err := sub.Hand(handData); err != nil {
						// rollback
						Log.Error().Msgf("[nats] init handle old data > %s > %s", dirname, err)
						if i > indexZero {
							clearCache(cache, int64(indexZero)-1, int64(i))
							dirname1, filename1 := sub.dirnames(since, i-1, c), sub.filenames(since, i-1, c)
							saveFastCache(cache, cacheDir, dirname1, filename1)
							_ = os.Remove(oldFile)
							_ = os.RemoveAll(filePath)
						}
						// reboot init handle old data
						time.Sleep(time.Second)
						sub.init(ctx)
						return
					}
					handRecords += len(handData)
					// reset data
					dataIndex = 0
					handData = [BulkSize][]byte{}
				}
			}
		}

		_ = os.Remove(oldFile)
		_ = os.RemoveAll(filePath)
	}

	Log.Info().Msgf("[nats] init handle old data > %d records", handRecords)
	if err := ctx.Err(); err != nil && err != context.Canceled {
		Log.Info().Msgf("[nats] init handle old data err > %s", err)
	}
}

// run handle new data.
func (sub *SubscriberFastCache) hand(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	var (
		runCount uint64
		delIndex uint64
	)

	var runHandle = func() {
		var handData [BulkSize][]byte
		count, handRecords := atomic.LoadUint64(&sub.Count), 0
		if count <= runCount {
			// reset handle
			if 0 < count && 3 == time.Now().Hour() && sub.Index <= delIndex {
				sub.Count, runCount, sub.Index, delIndex = 0, 0, 0, 0
			}
			return
		}

		for dataIndex := 0; dataIndex < BulkSize && sub.Index < count; runCount++ {
			key := atomic.AddUint64(&sub.Index, 1)
			handData[dataIndex] = sub.Cache.Get(nil, f.BytesUint64(key))
			handRecords++
			if dataIndex++; dataIndex == BulkSize || sub.Index == count {
				// bulk handle
				if err := sub.Hand(handData); err != nil {
					// rollback
					Log.Error().Msgf("[nats] run handle new data err > %s", err)
					sub.Index -= uint64(dataIndex)
					handRecords -= dataIndex
					break
				}
				// reset data
				dataIndex = 0
				handData = [BulkSize][]byte{}
			}
		}

		Log.Info().Msgf("[nats] run handle new data > %d records", handRecords)

		if handRecords == 0 {
			return
		}

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
				Log.Info().Msgf("[nats] done handle new data err > %s", err)
			}
			runHandle()
			return
		case <-time.After(time.Second):
			Log.Info().Msgf("[nats] run receive new data > %d/%d", sub.Index, sub.Count)
			runHandle()
		}
	}
}

func (sub *SubscriberFastCache) Dirname() string {
	return sub.dirname(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) dirname(since *f.TimeStamp, index, count uint64) string {
	return sub.dirnames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) dirnames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d", since, index, count)
}

func (sub *SubscriberFastCache) Filename() string {
	return sub.filename(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) filename(since *f.TimeStamp, index, count uint64) string {
	return sub.filenames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) filenames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d.json", since, index, count)
}

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
