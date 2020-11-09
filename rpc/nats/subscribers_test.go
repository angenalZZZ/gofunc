package nats_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/angenalZZZ/gofunc/f"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/nats-io/nats.go"
)

func TestSubscribers(t *testing.T) {
	// New Client Connect.
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())
	var num int32

	// Create subscriber list for Client Connect.
	subjList := []string{"TestSubscriber-1", "TestSubscriber-2", "TestSubscriber-3"}
	subjHand := func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			t.Logf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		}
		if i := atomic.AddInt32(&num, 1); i == 3 {
			f.DoneContext(ctx)
		}
	}
	handList := make([]nats.MsgHandler, 3)
	for i, _ := range subjList {
		handList[i] = subjHand
	}
	sub := nat.NewSubscribers(nat.Conn, subjList, handList)

	// Ping a message.
	go func() {
		time.Sleep(time.Millisecond)
		for _, s := range subjList {
			err = nat.Conn.Publish(s, []byte("ping"))
			if err != nil {
				t.Fatalf("[nats] failed publishing a %q message > %s", s, err.Error())
			} else {
				t.Logf("[nats] successful publishing a %q message", s)
			}
		}
	}()

	sub.Run(wait)
}
