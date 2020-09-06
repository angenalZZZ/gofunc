package nats

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"log"
	"syscall"
)

type Subscriber struct {
	*nats.Conn
	sub  *nats.Subscription
	Subj string
	Hand nats.MsgHandler
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
	var err error

	// Handle panic.
	defer func() {
		err := recover()
		if err != nil {
			Log.Error().Msgf("[nats] run error\t>\t%v", err)
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
	SubscribeLimitHandle(sub.sub, 10000000, 1048576)

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
		Log.Error().Msg("[nats] run forced termination")
	})
}
