package nats

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
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
	dir, since := "", f.TimeFrom(time.Now(), true)

	if len(cacheDir) == 1 && cacheDir[0] != "" {
		dir = filepath.Join(cacheDir[0], since.LocalTimeStampString(true))
	} else {
		dir = filepath.Join(f.CurrentDir(), ".nutsdb", since.LocalTimeStampString(true))
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
	handFunc := func(dir string) {
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

			handRecords, count := 0, len(items)
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
			handFunc(oldFile)
		}
	}

	Log.Info().Msgf("[nats] init handle old data > %d records", handRecords)
	if err := ctx.Err(); err != nil && err != context.Canceled {
		Log.Warn().Msgf("[nats] init handle old data err > %s", err)
	}
}
