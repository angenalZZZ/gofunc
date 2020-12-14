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

// Take time 770ms:60bytes, 4.52s:500bytes to process 1 million records,
// run times 1298700:60Bytes, 220848:500bytes Qps.(4CPU+16G+MHD)
func TestBenchmarkPublisher(t *testing.T) {
	// New Client Connect.
	nat.Subject = "BenchmarkPublisher"
	nat.Conn, err = nat.New("nats.go", "", "", "HGJ766GR767FKJU0", "", "")
	if err != nil {
		t.Fatalf("[nats] failed to connect: %s\n", err.Error())
	}

	var publishedNumber, succeededNumber, failedNumber int64

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriber(nat.Conn, nat.Subject, func(msg *nats.Msg) {
		atomic.AddInt64(&succeededNumber, 1)
	})

	ctx, wait := f.ContextWithWait(context.TODO())
	go sub.Run(wait)
	time.Sleep(time.Millisecond)

	var bufferData = random.AlphaNumberBytes(500)

	// start benchmark test
	t1 := time.Now()

	// test publish pressure
	for i := 0; i < 1000000; i++ {
		err = nat.Conn.Publish(sub.Subj, bufferData)
		if err != nil {
			failedNumber++
		} else {
			publishedNumber++
		}
	}

	// wait succeeded-number equals published-number
	f.NumIncrWait(&publishedNumber, &succeededNumber)
	t2 := time.Now()
	ts := t2.Sub(t1)
	//time.Sleep(time.Millisecond)
	f.DoneContext(ctx)

	t.Logf("Publish Number: %d, Successful Number: %d, Failed Number %d", publishedNumber, succeededNumber, failedNumber)
	t.Logf("Take time %s, handle received messages %d qps", ts, 1000*succeededNumber/ts.Milliseconds())
}
