package nats

import (
	"context"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClientConnect() (nc *nats.Conn, err error) {
	return New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
}

func TestSubscriber(t *testing.T) {
	// New Client Connect.
	nc, err := newTestClientConnect()
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := NewSubscriber(nc, "TestSubscriber", func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			t.Logf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		}
		f.DoneContext(ctx)
	})

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

func BenchmarkPublisher(b *testing.B) {
	// New Client Connect.
	nc, err := newTestClientConnect()
	if err != nil {
		b.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	var publishedNumber, succeededNumber, failedNumber int64

	// Create a subscriber for Client Connect.
	sub := NewSubscriber(nc, "BenchmarkPublisher", func(msg *nats.Msg) {
		atomic.AddInt64(&succeededNumber, 1)
	})

	ctx, wait := f.ContextWithWait(context.TODO())
	go sub.Run(wait)
	time.Sleep(time.Millisecond)

	var bufferData = random.AlphaNumberBytes(60)

	// start benchmark test
	b.ResetTimer()

	// test publish pressure
	for i := 0; i < b.N; i++ {
		err = nc.Publish(sub.Subj, bufferData)
		if err != nil {
			atomic.AddInt64(&failedNumber, 1)
		} else {
			atomic.AddInt64(&publishedNumber, 1)
		}
	}

	// wait succeeded-number equals published-number
	f.NumIncrWait(&publishedNumber, &succeededNumber)
	f.DoneContext(ctx)
	b.StopTimer()

	b.Logf("Publish Number: %d, Successful Number: %d, Failed Number %d", publishedNumber, succeededNumber, failedNumber)
}
