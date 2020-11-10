package nats

import (
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
	Hand         func([]string) error
	Cache        f.CMap
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
		Cache:        f.NewConcurrentMap(), // fast Concurrent Map
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
