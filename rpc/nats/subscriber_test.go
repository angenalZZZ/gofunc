package nats

import (
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestNewSubscriber(t *testing.T) {
	// New Client Connect.
	nc, err := New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	// Create a subscriber for Client Connect.
	sub := NewSubscriber(nc, "OpLogCommand", func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			t.Logf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		}
	})

	// Ping a message.
	go func() {
		time.Sleep(time.Second)
		err = nc.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			t.Fatalf("[nats] failed publishing a test message\t>\t%s", err.Error())
		}
	}()

	sub.Run(3 * time.Second)
}
