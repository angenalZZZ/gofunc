package nats

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"log"
	"syscall"
)

type Subscriber struct {
	*nats.Conn
	Hand nats.MsgHandler
	Subj string
}

// NewSubscriber Create a subscriber for Client Connect.
func NewSubscriber(nc *nats.Conn, subject string, msgHandler nats.MsgHandler) *Subscriber {
	sub := &Subscriber{
		Conn: nc,
		Subj: subject,
		Hand: msgHandler,
	}
	return sub
}

// Run runtime to end your application.
func (sub *Subscriber) Run(waitFunc ...func()) {
	hasWait := len(waitFunc) > 0

	// Handle panic
	defer func() {
		if err := recover(); err != nil {
			Log.Error().Msgf("[nats] run error\t>\t%s", err)
			log.Panic(err)
		} else if hasWait {
			// Drain connection (Preferred for responders), Close() not needed if this is called.
			if err = sub.Conn.Drain(); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// Async Subscriber
	s, err := sub.Conn.Subscribe(sub.Subj, sub.Hand)
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
