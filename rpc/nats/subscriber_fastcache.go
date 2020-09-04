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

	// init handle old data
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

// init handle old data
func (sub *SubscriberFastCache) init() {
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

		var data [CacheBulkSize][]byte
		for i, c, dataIndex := indexZero, uint64(count), 0; i <= c; i++ {
			if key := f.BytesUint64(i); cache.Has(key) {
				data[dataIndex] = cache.Get(nil, key)
				if dataIndex++; dataIndex == CacheBulkSize || i == c {
					// bulk handle
					if err := sub.Hand(data); err != nil {
						// rollback
						Log.Error().Msgf("[nats] init handle old data\t>\t%s", err)
						if i > indexZero {
							clearCache(cache, int64(indexZero)-1, int64(i))
							dirname1, filename1 := sub.dirnames(since, i-1, c), sub.filenames(since, i-1, c)
							saveFastCache(cache, cacheDir, dirname1, filename1)
							_ = os.Remove(oldFile)
							_ = os.RemoveAll(filePath)
						}
						// reboot init handle old data
						time.Sleep(time.Second)
						sub.init()
						return
					}
					// reset data
					dataIndex = 0
					data = [CacheBulkSize][]byte{}
				}
			}
		}

		_ = os.Remove(oldFile)
		_ = os.RemoveAll(filePath)
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
