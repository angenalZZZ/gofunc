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

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
)

// SubscriberFast The NatS Subscriber with Fast Concurrent Map Temporary Storage.
type SubscriberFast struct {
	*nats.Conn
	sub          *nats.Subscription
	Subj         string
	Hand         func([][]byte) error
	Cache        f.CiMap
	CacheDir     string // sets cache persist to disk directory
	Index        uint64
	Count        uint64
	Since        *f.TimeStamp
	MsgLimit     int   // sets the limits for pending messages for this subscription.
	BytesLimit   int   // sets the limits for a message's bytes for this subscription.
	OnceAmount   int64 // sets amount allocated at one time
	OnceInterval time.Duration
	async        bool
	err          error
}

// NewSubscriberFast Create a subscriber with cache store for Client Connect.
func NewSubscriberFast(nc *nats.Conn, subject string, cacheDir ...string) *SubscriberFast {
	sub := &SubscriberFast{
		Conn:         nc,
		Subj:         subject,
		Cache:        f.NewCiMap(), // fast Concurrent Map
		Since:        f.TimeFrom(time.Now(), true),
		MsgLimit:     100000000, // pending messages: 100 million
		BytesLimit:   1048576,   // a message's size: 1MB
		OnceAmount:   -1,
		OnceInterval: time.Second,
		async:        true,
	}
	if len(cacheDir) == 1 && cacheDir[0] != "" {
		sub.CacheDir = cacheDir[0]
	} else {
		sub.CacheDir = data.CurrentDir
	}
	return sub
}

// LimitMessage sets amount for pending messages for this subscription, and a message's bytes.
// Defaults amountPendingMessages: 100 million, anMessageBytes: 1MB
func (sub *SubscriberFast) LimitMessage(amountPendingMessages, anMessageBytes int) {
	sub.MsgLimit, sub.BytesLimit = amountPendingMessages, anMessageBytes
}

// LimitAmount sets amount allocated at one time, and the processing interval time.
// Defaults onceAmount: -1, onceInterval: time.Second
func (sub *SubscriberFast) LimitAmount(onceAmount int64, onceInterval time.Duration) {
	sub.OnceAmount, sub.OnceInterval = onceAmount, onceInterval
}

// Run runtime to end your application.
func (sub *SubscriberFast) Run(waitFunc ...func()) {
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
		// Save cache.
		sub.Save(sub.CacheDir)

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber.
	sub.sub, sub.err = sub.Conn.Subscribe(sub.Subj, func(msg *nats.Msg) {
		key, val := atomic.AddUint64(&sub.Count, 1), msg.Data
		sub.Cache.Set(key, val)
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
func (sub *SubscriberFast) init(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	oldFiles, _ := filepath.Glob(filepath.Join(sub.CacheDir, "*"))
	sort.Strings(oldFiles)

	cacheDir := sub.CacheDir
	var clearCache = func(cache f.CiMap, index, count int64) {
		for i, c := uint64(index)+1, uint64(count); i <= c; i++ {
			cache.Remove(i)
		}
	}

	handRecords, onceRecords := 0, atomic.LoadInt64(&sub.OnceAmount)
	for _, oldFile := range oldFiles {
		_, jsonFile := filepath.Split(oldFile)
		if ok, _ := regexp.MatchString(`^\d+\.\d+\.\d+\.json$`, jsonFile); !ok {
			continue
		}

		dirname := strings.ReplaceAll(jsonFile, ".json", "")
		fileData, err := ioutil.ReadFile(oldFile)
		if err != nil {
			continue
		}

		cache, err := f.NewCiMapFromJSON(fileData)
		s := strings.Split(dirname, ".")
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
			if val, ok := cache.Get(i); ok {
				handData = append(handData, val.([]byte))
				if dataIndex++; dataIndex == onceRecords || i == c {
					// bulk handle
					if err := sub.Hand(handData); err != nil {
						// rollback
						Log.Error().Msgf("[nats] init handle old data > %s > %s", dirname, err)
						if i > indexZero {
							clearCache(cache, int64(indexZero)-1, int64(i))
							dirname1, filename1 := sub.dirnames(since, i-1, c), sub.filenames(since, i-1, c)
							saveFastMap(cache, cacheDir, dirname1, filename1)
							_ = os.Remove(oldFile)
						}
						// reboot init handle old data
						time.Sleep(sub.OnceInterval)
						sub.init(ctx)
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
	}

	Log.Info().Msgf("[nats] init handle old data > %d records", handRecords)
	if err := ctx.Err(); err != nil && err != context.Canceled {
		Log.Warn().Msgf("[nats] init handle old data err > %s", err)
	}
}

// run handle new data.
func (sub *SubscriberFast) hand(ctx context.Context) {
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
			val, ok := sub.Cache.Get(index)
			if !ok || val == nil {
				val = []byte{}
			}
			handData = append(handData, val.([]byte))
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
				sub.Cache.Remove(i)
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
func (sub *SubscriberFast) Dirname() string {
	return sub.dirname(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFast) dirname(since *f.TimeStamp, index, count uint64) string {
	return sub.dirnames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFast) dirnames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d", since, index, count)
}

// Filename gets the Cache Filename.
func (sub *SubscriberFast) Filename() string {
	return sub.filename(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFast) filename(since *f.TimeStamp, index, count uint64) string {
	return sub.filenames(since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFast) filenames(since string, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d.json", since, index, count)
}

// Save the Cache Data of some records not processed.
func (sub *SubscriberFast) Save(cacheDir string) {
	if sub.Count == 0 || sub.Hand == nil || sub.Index == sub.Count {
		return
	}

	saveFastMap(sub.Cache, cacheDir, sub.Dirname(), sub.Filename())
}

func saveFastMap(cache f.CiMap, cacheDir, dirname, filename string) {
	handData, err := cache.JSON()
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats > %s", err)
	}

	filePath := filepath.Join(cacheDir, filename)
	err = ioutil.WriteFile(filePath, handData, 0644)
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats > %s", err)
	}
}
