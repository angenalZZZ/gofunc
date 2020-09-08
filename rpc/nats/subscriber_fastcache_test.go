package nats

import (
	"context"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	"testing"
	"time"
)

func TestSubscriberFastCache(t *testing.T) {
	// New Client Connect.
	nc, err := newTestClientConnect()
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := NewSubscriberFastCache(nc, "TestSubscriberFastCache", data.RootDir)
	sub.Hand = func(list [HandSize][]byte) error {
		for _, item := range list {
			if len(item) == 0 {
				break
			}
			if item[0] != '{' {
				t.Logf("[nats] received test message on %q: %s", sub.Subj, string(item))
			}
		}

		f.DoneContext(ctx)
		t.Logf("[nats] test finished on %q", sub.Subj)
		return nil
	}

	// Ping a message.
	go func() {
		time.Sleep(time.Millisecond)
		err = nc.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			t.Fatalf("[nats] failed publishing a test message\t>\t%s", err.Error())
		} else {
			t.Logf("[nats] successful publishing a test message")
		}
	}()

	sub.Run(wait)
}
