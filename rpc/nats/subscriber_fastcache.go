package nats

import (
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"log"
	"sync/atomic"
	"syscall"
)

type SubscriberFastCache struct {
	*nats.Conn
	*fastcache.Cache
	Subj string
	key  uint64
}

// SubscriberFastCache Create a subscriber with cache store for Client Connect.
func NewSubscriberFastCache(nc *nats.Conn, subject string) *SubscriberFastCache {
	sub := &SubscriberFastCache{
		Conn:  nc,
		Subj:  subject,
		Cache: fastcache.New(2048),
		key:   1,
	}
	return sub
}

// Run runtime to end your application.
func (sub *SubscriberFastCache) Run(waitFunc ...func()) {
	hasWait := len(waitFunc) > 0

	// Handle panic
	defer func() {
		if err := recover(); err != nil {
			_ = sub.Cache.SaveToFileConcurrent(f.CurrentDir()+"/"+sub.Subj, 4)
			Log.Error().Msgf("[nats] run error\t>\t%s", err)
			log.Panic(err)
		} else if hasWait {
			_ = sub.Cache.SaveToFileConcurrent(f.CurrentDir()+"/"+sub.Subj, 4)
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
			k := atomic.AddUint64(&sub.key, 1)
			sub.Cache.Set(f.BytesUint64(k), msg.Data)
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

	if hasWait {
		waitFunc[0]()
		return
	}

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		if err = sub.Conn.Drain(); err != nil {
			log.Fatal(err)
		}
	})
}
