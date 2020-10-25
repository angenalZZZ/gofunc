package nats

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"log"
	"syscall"
	"time"
)

type Subscriber struct {
	*nats.Conn
	sub        *nats.Subscription
	Subj       string
	Hand       nats.MsgHandler
	Since      *f.TimeStamp
	MsgLimit   int // sets the limits for pending messages for this subscription.
	BytesLimit int // sets the limits for a message's bytes for this subscription.
}

// NewSubscriber Create a subscriber for Client Connect.
func NewSubscriber(nc *nats.Conn, subject string, msgHandler nats.MsgHandler) *Subscriber {
	sub := &Subscriber{
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
func (sub *Subscriber) Run(waitFunc ...func()) {
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
		_ = sub.sub.Unsubscribe()
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = sub.Conn.Drain()

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber.
	sub.sub, err = sub.Conn.Subscribe(sub.Subj, sub.Hand)
	// Set listening.
	SubscribeErrorHandle(sub.sub, true, err)
	if err != nil {
		log.Fatal(err)
	}

	// Set pending limits.
	SubscribeLimitHandle(sub.sub, sub.MsgLimit, sub.BytesLimit)

	// Flush connection to server, returns when all messages have been processed.
	FlushAndCheckLastError(sub.Conn)

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
