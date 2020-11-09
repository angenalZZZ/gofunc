package nats

import (
	"log"
	"syscall"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
)

// Subscribers subscriber list for a Client Connect.
type Subscribers struct {
	*nats.Conn
	sub        []*nats.Subscription
	Subj       []string
	Hand       []nats.MsgHandler
	Since      *f.TimeStamp
	MsgLimit   int // sets the limits for pending messages for this subscription.
	BytesLimit int // sets the limits for a message's bytes for this subscription.
}

// NewSubscribers Create subscriber list for a Client Connect.
func NewSubscribers(nc *nats.Conn, subject []string, msgHandler []nats.MsgHandler) *Subscribers {
	sub := &Subscribers{
		Conn:       nc,
		Subj:       subject,
		Hand:       msgHandler,
		Since:      f.TimeFrom(time.Now(), true),
		MsgLimit:   100000000, // pending messages: 100 million
		BytesLimit: 1048576,   // a message's size: 1MB
	}
	return sub
}

// Run runtime to end your application.
func (sub *Subscribers) Run(waitFunc ...func()) {
	var err error

	// Handle panic.
	defer func() {
		err := recover()
		if err != nil {
			Log.Error().Msgf("[nats] run error > %v", err)
		} else {
			Log.Warn().Msg("[nats] stop receive new data")
		}

		// Unsubscribe will remove interest in the given subject.
		for _, s := range sub.sub {
			_ = s.Unsubscribe()
		}
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = sub.Conn.Drain()

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber.
	sub.sub = make([]*nats.Subscription, len(sub.Subj))
	for i, subj := range sub.Subj {
		sub.sub[i], err = sub.Conn.Subscribe(subj, sub.Hand[i])

		// Set listening.
		SubscribeErrorHandle(sub.sub[i], true, err)
		if err != nil {
			log.Fatal(err)
		}

		// Set pending limits.
		SubscribeLimitHandle(sub.sub[i], sub.MsgLimit, sub.BytesLimit)

		// Flush connection to server, returns when all messages have been processed.
		FlushAndCheckLastError(sub.Conn)
	}

	if len(waitFunc) > 0 {
		waitFunc[0]()
		return
	}

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		Log.Error().Msg("[nats] forced to shutdown.")
	})
}
