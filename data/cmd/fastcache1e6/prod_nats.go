package main

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"syscall"
)

func ProdNatS() {
	var (
		name = "cache.>"
		sub  *nats.Subscription
	)

	// "*" matches any token, at any level of the subject.
	// ">" matches any length of the tail of a subject, and can only be the last token
	// E.g. 'cache.>' will match 'cache.set.123', 'cache.get.123', 'cache.del.123'
	if *flagName != "" {
		name = *flagName + ".>"
	}

	nc, err := nat.New("fastcache1e6", *flagAddr, "", *flagToken, *flagCert, *flagKey)

	if err != nil {
		nat.Log.Error().Msgf("Nats failed connect to server: %v\n", err)
		os.Exit(1)
	}

	// Handle panic.
	defer func() {
		err := recover()
		if err != nil {
			nat.Log.Error().Msgf("[nats] run error > %v", err)
		}

		// Unsubscribe will remove interest in the given subject.
		_ = sub.Unsubscribe()
		// Drain connection (Preferred for responders), Close() not needed if this is called.
		_ = nc.Drain()

		// os.Exit(1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Async Subscriber.
	sub, err = nc.Subscribe(name, func(m *nats.Msg) {
		result := defaultService.Handle(m.Data)
		if err = m.Respond(result); err != nil {
			_ = fmt.Errorf("Nats failed to Write: %v\n", err)
		}
		//if err = nc.Publish(m.Reply, result); err != nil {
		//	_ = fmt.Errorf("Nats failed to Write: %v\n", err)
		//}
	})
	// Set listening.
	nat.SubscribeErrorHandle(sub, true, err)
	if err != nil {
		os.Exit(1)
	}

	// Set pending limits.
	nat.SubscribeLimitHandle(sub, 10000000, 1048576)

	// Flush connection to server, returns when all messages have been processed.
	nat.FlushAndCheckLastError(nc)

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		nat.Log.Error().Msg("[nats] run forced termination")
	})
}
