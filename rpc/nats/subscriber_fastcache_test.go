package nats_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

func TestSubscriberFastCache(t *testing.T) {
	// New Client Connect.
	nat.Subject = "TestSubscriberFastCache"
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriberFastCache(nat.Conn, nat.Subject, data.RootDir)
	sub.Hand = func(list [][]byte) error {
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
		err = nat.Conn.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			t.Fatalf("[nats] failed publishing a test message\t>\t%s", err.Error())
		} else {
			t.Logf("[nats] successful publishing a test message")
		}
	}()

	sub.Run(wait)
}

func TestBenchSubscriberFastCache(t *testing.T) {
	// New Client Connect.
	nat.Subject = "BenchmarkSubscriberFastCache"
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	var publishedNumber, succeededNumber, failedNumber int64

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriberFastCache(nat.Conn, nat.Subject, data.RootDir)
	sub.Hand = func(list [][]byte) error {
		var n int64
		for _, item := range list {
			//Log.Info().Msgf("[nats] received message on %q: %s", sub.Subj, string(item))
			if len(item) == 0 {
				break
			}
			n++
		}

		atomic.AddInt64(&succeededNumber, n)
		return nil
	}

	ctx, wait := f.ContextWithWait(context.TODO())
	go sub.Run(wait)
	time.Sleep(time.Millisecond)

	var bufferData = random.AlphaNumberBytes(60)

	// start benchmark test
	t1 := time.Now()

	// test publish pressure
	for i := 0; i < 1000000; i++ {
		err = nat.Conn.Publish(sub.Subj, bufferData)
		if err != nil {
			atomic.AddInt64(&failedNumber, 1)
		} else {
			atomic.AddInt64(&publishedNumber, 1)
		}
	}

	// wait succeeded-number equals published-number
	f.NumIncrWait(&publishedNumber, &succeededNumber)
	t2 := time.Now()
	ts := t2.Sub(t1)
	time.Sleep(time.Millisecond)
	f.DoneContext(ctx)

	t.Logf("Publish Number: %d, Successful Number: %d, Failed Number %d", publishedNumber, succeededNumber, failedNumber)
	t.Logf("Take time %s, handle received messages %d qps", ts, 1000*succeededNumber/ts.Milliseconds())
}
