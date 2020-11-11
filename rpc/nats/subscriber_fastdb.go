package nats

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"github.com/xujiajun/nutsdb"
)

// SubscriberFastDb The NatS Subscriber with Fast DB Storage.
type SubscriberFastDb struct {
	*nats.Conn
	sub          *nats.Subscription
	Subj         string
	Hand         func([][]byte) error
	Cache        *nutsdb.DB
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

// NewSubscriberFastDb Create a subscriber with cache store for Client Connect.
func NewSubscriberFastDb(nc *nats.Conn, subject string, cacheDir ...string) *SubscriberFastDb {
	dir, since := ".nutsdb", f.TimeFrom(time.Now(), true)

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

	opt := newNdbOptions(dir)
	client, _ := nutsdb.Open(opt)

	sub := &SubscriberFastDb{
		Conn:         nc,
		Subj:         subject,
		Cache:        client, // fast nutsdb
		CacheDir:     opt.Dir,
		Since:        since,
		MsgLimit:     100000000, // pending messages: 100 million
		BytesLimit:   1048576,   // a message's size: 1MB
		OnceAmount:   -1,
		OnceInterval: time.Second,
		async:        true,
	}

	return sub
}

func newNdbOptions(dir string) nutsdb.Options {
	opt := nutsdb.DefaultOptions
	opt.Dir = dir
	opt.SyncEnable = false
	return opt
}

// LimitMessage sets amount for pending messages for this subscription, and a message's bytes.
// Defaults amountPendingMessages: 100 million, anMessageBytes: 1MB
func (sub *SubscriberFastDb) LimitMessage(amountPendingMessages, anMessageBytes int) {
	sub.MsgLimit, sub.BytesLimit = amountPendingMessages, anMessageBytes
}

// LimitAmount sets amount allocated at one time, and the processing interval time.
// Defaults onceAmount: -1, onceInterval: time.Second
func (sub *SubscriberFastDb) LimitAmount(onceAmount int64, onceInterval time.Duration) {
	sub.OnceAmount, sub.OnceInterval = onceAmount, onceInterval
}

// Run runtime to end your application.
func (sub *SubscriberFastDb) Run(waitFunc ...func()) {
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
		// Stop cache processor.
		_ = sub.Cache.Close()

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber.
	sub.sub, sub.err = sub.Conn.Subscribe(sub.Subj, func(msg *nats.Msg) {
		key, val := atomic.AddUint64(&sub.Count, 1), msg.Data
		_ = sub.Cache.Update(func(tx *nutsdb.Tx) (err error) {
			return tx.Put(sub.Subj, f.BytesUint64(key), val, 0)
		})
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
func (sub *SubscriberFastDb) init(ctx context.Context) {
	if sub.Hand == nil {
		return
	}

	bucket, cacheDir := sub.Subj, sub.CacheDir
	dir0, parentDir := filepath.Base(cacheDir), filepath.Dir(cacheDir)
	oldFiles, _ := filepath.Glob(filepath.Join(parentDir, "*"))
	sort.Strings(oldFiles)

	var clearCache = func(cache *nutsdb.DB, keys [][]byte) {
		for _, key := range keys {
			_ = cache.Update(func(tx *nutsdb.Tx) error {
				return tx.Delete(bucket, key)
			})
		}
	}

	handRecords, onceRecords := 0, atomic.LoadInt64(&sub.OnceAmount)
	var runHandle = func(dir string) {
		opt := newNdbOptions(dir)
		client, err := nutsdb.Open(opt)
		if err != nil {
			return
		}

		var ok bool
		dirname, keys := filepath.Base(dir), make([][]byte, 0, 1000)
		defer func() {
			_ = client.Close()
			if ok {
				_ = os.RemoveAll(dir)
			}
		}()

		err = client.View(func(tx *nutsdb.Tx) error {
			items, err := tx.GetAll(bucket)
			if err != nil {
				return err
			}

			count := len(items)
			indexSize := count
			if onceRecords > 0 {
				indexSize = int(onceRecords)
			}

			handData, handKeys := make([][]byte, 0, indexSize), make([][]byte, 0, indexSize)

			for i, c, dataIndex := 0, count, int64(0); i <= c; i++ {
				if val := items[i].Value; val != nil {
					handData = append(handData, val)
					handKeys = append(handKeys, items[i].Key)
					if dataIndex++; dataIndex == onceRecords || i == c {
						// bulk handle
						if err := sub.Hand(handData); err != nil {
							Log.Error().Msgf("[nats] init handle old data > %s > %s", dirname, err)
							return err
						}
						handRecords += len(handData)
						for _, key := range handKeys {
							keys = append(keys, key)
						}
						// reset data
						dataIndex = 0
						handData = make([][]byte, 0, indexSize)
						time.Sleep(sub.OnceInterval)
					}
				}
			}

			return nil
		})

		if len(keys) > 0 {
			clearCache(client, keys)
		}
		if err != nil {
			return
		}

		ok = true
	}

	for _, oldFile := range oldFiles {
		if !f.IsDir(oldFile) {
			continue
		}

		dir1 := filepath.Base(oldFile)
		if dir1 == dir0 {
			continue
		}

		// reboot init handle old data
		for f.PathExists(oldFile) {
			runHandle(oldFile)
			if handRecords > 0 {
				Log.Info().Msgf("[nats] init handle old data > %d records", handRecords)
			}
		}
	}

	if err := ctx.Err(); err != nil && err != context.Canceled {
		Log.Warn().Msgf("[nats] init handle old data err > %s", err)
	}
}

// run handle new data.
func (sub *SubscriberFastDb) hand(ctx context.Context) {
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
			var val []byte
			err := sub.Cache.View(func(tx *nutsdb.Tx) error {
				item, err := tx.Get(sub.Subj, f.BytesUint64(index))
				if err != nil {
					return err
				}
				val = item.Value
				return nil
			})
			if err != nil || val == nil {
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
				_ = sub.Cache.Update(func(tx *nutsdb.Tx) error {
					return tx.Delete(sub.Subj, f.BytesUint64(i))
				})
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
