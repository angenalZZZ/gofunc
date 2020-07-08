package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func ProdNatS() {
	//addr := fmt.Sprintf("nats://127.0.0.1:%d", *flagPort)
	addr, name := nats.DefaultURL, *flagName+".>"

	nc, err := nats.Connect(addr, nats.Token(*flagToken))
	if err != nil {
		_ = fmt.Errorf("Nats failed connect to server: %v\n", err)
		return
	}
	defer func() {
		// Flush connection to server, returns when all messages have been processed.
		_ = nc.Flush()
		// Drain connection (Preferred for responders)
		// Close() not needed if this is called.
		_ = nc.Drain()
		// Close connection
		nc.Close()
	}()
	fmt.Printf("Nats client connected to %s\n", addr)
	for {
		_, err := nc.Subscribe(name, func(m *nats.Msg) {
			result := defaultService.Handle(m.Data)
			if err = nc.Publish(m.Reply, result); err != nil {
				_ = fmt.Errorf("Nats failed to Write: %v\n", err)
			}
		})
		if err != nil {
			_ = fmt.Errorf("Nats failed to Read: %v\n", err)
			return
		}
	}
}
