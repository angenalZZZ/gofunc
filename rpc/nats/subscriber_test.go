package nats

import (
	"context"
	"fmt"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/nats-io/nats.go"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewSubscriber(t *testing.T) {
	// New Client Connect.
	nc, err := New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := NewSubscriber(nc, "OpLogCommand", func(msg *nats.Msg) {
		if msg.Data[0] != '{' {
			t.Logf("[nats] received test message on %q: %s", msg.Subject, string(msg.Data))
		}
		f.DoneContext(ctx)
	})

	// Ping a message.
	go func() {
		time.Sleep(time.Second)
		err = nc.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			t.Fatalf("[nats] failed publishing a test message\t>\t%s", err.Error())
		} else {
			t.Logf("[nats] successful publishing a test message")
		}
	}()

	sub.Run(wait)
}

func BenchmarkNewSubscriber(b *testing.B) {
	var (
		publishNumber, succeededNumber, failedNumber int64
		bufferData                                   = random.AlphaNumberBytes(60)
	)

	// New Client Connect.
	nc, err := New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		b.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	// Create a subscriber for Client Connect.
	sub := NewSubscriber(nc, "OpLogCommand", func(msg *nats.Msg) {
		atomic.AddInt64(&succeededNumber, 1)
	})

	ctx, wait := f.ContextWithWait(context.Background())
	sub.Run(wait)

	for _, concurrency := range []int{1, 2, 4, 8, 16, 24, 32} {
		b.Run(fmt.Sprintf("concurrency_%d", concurrency), func(b *testing.B) {
			// Start benchmark
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				err = nc.Publish(sub.Subj, bufferData)
				if err != nil {
					atomic.AddInt64(&failedNumber, 1)
				} else {
					atomic.AddInt64(&publishNumber, 1)
				}
			}
		})
	}

	if atomic.LoadInt64(&succeededNumber) <= atomic.LoadInt64(&publishNumber) {
		time.Sleep(time.Millisecond)
	}
	f.DoneContext(ctx)
	b.Logf("Successful Number: %d, Failed Number %d", succeededNumber, failedNumber)
}
