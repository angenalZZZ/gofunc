package nats_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/nats-io/nats.go"
)

var err error

func TestSubscriber(t *testing.T) {
	// New Client Connect.
	nat.Subject = "TestSubscriber"
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriber(nat.Conn, nat.Subject, func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			t.Logf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		}
		f.DoneContext(ctx)
	})

	// Ping a message.
	go func() {
		time.Sleep(time.Millisecond)
		err = nat.Conn.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			t.Fatalf("[nats] failed publishing a test message > %s", err.Error())
		} else {
			t.Logf("[nats] successful publishing a test message")
		}
	}()

	sub.Run(wait)
}

func BenchmarkPublisher(b *testing.B) {
	// New Client Connect.
	nat.Subject = "BenchmarkPublisher"
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		b.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	var publishedNumber, succeededNumber, failedNumber int64

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriber(nat.Conn, nat.Subject, func(msg *nats.Msg) {
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
		err = nat.Conn.Publish(sub.Subj, bufferData)
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
