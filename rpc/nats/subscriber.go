package nats

import (
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"syscall"
)

type Subscriber struct {
	*nats.Conn
	nats.MsgHandler
	Subject string
}

func NewSubscriber(nc *nats.Conn, subject string) *Subscriber {
	sub := &Subscriber{
		Conn:    nc,
		Subject: subject,
	}
	return sub
}

func (sub *Subscriber) Run() {
	defer func() {
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = sub.Conn.Drain()
	}()

	// async Subscriber
	s, err := sub.Conn.Subscribe(sub.Subject, sub.MsgHandler)
	SubscribeErrorHandle(s, true, err)
	SubscribeLimitHandle(s, 10000000, 1048576)

	// Flush connection to server, returns when all messages have been processed.
	FlushAndCheckLastError(sub.Conn)

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}
