package nats

import (
	"fmt"
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

const CacheBulkSize int = 1000

type SubscriberFastCache struct {
	*nats.Conn
	Subj string
	Hand func([CacheBulkSize][]byte) error
	*fastcache.Cache
	CacheDir string // cache persist to disk directory
	Index    uint64
	Count    uint64
	Since    *f.TimeStamp
	saved    bool
}

// SubscriberFastCache Create a subscriber with cache store for Client Connect.
func NewSubscriberFastCache(nc *nats.Conn, subject string, cacheDir ...string) *SubscriberFastCache {
	sub := &SubscriberFastCache{
		Conn:  nc,
		Subj:  subject,
		Cache: fastcache.New(2048),
		Since: f.TimeFrom(time.Now(), true),
	}
	if len(cacheDir) == 1 && cacheDir[0] != "" {
		sub.CacheDir = cacheDir[0]
	} else {
		sub.CacheDir = f.CurrentDir()
	}
	return sub
}

// Run runtime to end your application.
func (sub *SubscriberFastCache) Run(waitFunc ...func()) {
	hasWait := len(waitFunc) > 0

	// Handle panic
	defer func() {
		if err := recover(); err != nil {
			sub.Save(sub.CacheDir)
			Log.Error().Msgf("[nats] run error\t>\t%s", err)
			log.Panic(err)
		} else if hasWait {
			sub.Save(sub.CacheDir)
			// Drain connection (Preferred for responders), Close() not needed if this is called.
			if err = sub.Conn.Drain(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// Async Subscriber
	s, err := sub.Conn.Subscribe(sub.Subj, func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			Log.Info().Msgf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		} else {
			key := atomic.AddUint64(&sub.Count, 1)
			sub.Cache.Set(f.BytesUint64(key), msg.Data)
		}
	})
	SubscribeErrorHandle(s, true, err)
	if err != nil {
		panic(err)
	}

	// Set pending limits
	SubscribeLimitHandle(s, 10000000, 1048576)

	// Flush connection to server, returns when all messages have been processed.
	FlushAndCheckLastError(sub.Conn)

	// Init handle old data
	go sub.init()

	// Todo handle new data

	if hasWait {
		waitFunc[0]()
		return
	}

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		sub.Save(sub.CacheDir)
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		if err = sub.Conn.Drain(); err != nil {
			log.Fatal(err)
		}
	})
}

func (sub *SubscriberFastCache) init() {
	oldFiles, _ := filepath.Glob(filepath.Join(sub.CacheDir, "*"))
	sort.Strings(oldFiles)

	cacheDir := sub.CacheDir
	var clearCache = func(cache *fastcache.Cache, index, count int64) {
		for i, c := uint64(index)+1, uint64(count); i <= c; i++ {
			k := f.BytesUint64(i)
			cache.Del(k)
		}
	}

	for _, oldFile := range oldFiles {
		dir, jsonFile := filepath.Split(oldFile)
		if ok, _ := regexp.MatchString(`^\d+\.\d+\.\d+\.json$`, jsonFile); !ok {
			continue
		}

		dirname := strings.ReplaceAll(jsonFile, ".json", "")
		filePath := filepath.Join(dir, dirname)
		cache, err := fastcache.LoadFromFile(filePath)
		s := strings.Split(dirname, ".")
		if err != nil || len(s) != 3 {
			continue
		}

		index, _ := strconv.ParseInt(s[1], 10, 0)
		count, _ := strconv.ParseInt(s[2], 10, 0)
		if index == 0 || count == 0 {
			continue
		}

		var data [CacheBulkSize][]byte
		for i, c, x := uint64(index)+1, uint64(count), 0; i <= c; i++ {
			if key := f.BytesUint64(i); cache.Has(key) {
				data[x] = cache.Get(nil, key)
				x++
				if x == CacheBulkSize || i+1 == c {
					x = 0
					// bulk handle data
					if sub.Hand != nil && sub.Hand(data) != nil {
						// rollback data
						index = int64(i)
						clearCache(cache, 0, index)
						since1 := f.TimeFrom(time.Now(), true)
						dirname1, filename1 := sub.Dirnames(since1, i, c), sub.Filenames(since1, i, c)
						saveFastCache(cache, cacheDir, dirname1, filename1)
						_ = os.Remove(oldFile)
						_ = os.RemoveAll(filePath)
						return
					}
				}
			}
		}

		_ = os.Remove(oldFile)
		_ = os.RemoveAll(filePath)
	}
}

func (sub *SubscriberFastCache) Dirname() string {
	return sub.Dirnames(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) Dirnames(since *f.TimeStamp, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d", since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) Filename() string {
	return sub.Filenames(sub.Since, sub.Index, sub.Count)
}

func (sub *SubscriberFastCache) Filenames(since *f.TimeStamp, index, count uint64) string {
	return fmt.Sprintf("%s.%d.%d.json", since.LocalTimeStampString(true), index, count)
}

func (sub *SubscriberFastCache) Save(cacheDir string) {
	if sub.saved || sub.Count == 0 || sub.Hand == nil {
		return
	}

	sub.saved = true
	saveFastCache(sub.Cache, cacheDir, sub.Dirname(), sub.Filename())
}

func saveFastCache(cache *fastcache.Cache, cacheDir, dirname, filename string) {
	fileStat := new(fastcache.Stats)
	cache.UpdateStats(fileStat)
	data, err := f.EncodeJson(fileStat)
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats\t>\t%s", err)
	}

	filePath := filepath.Join(cacheDir, filename)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		Log.Error().Msgf("[nats] save cache stats\t>\t%s", err)
	}

	dirPath := filepath.Join(cacheDir, dirname)
	if err = cache.SaveToFileConcurrent(dirPath, 0); err != nil {
		Log.Error().Msgf("[nats] save cache data\t>\t%s", err)
	} else {
		cache.Reset() // Reset removes all the items from the cache.
	}
}
