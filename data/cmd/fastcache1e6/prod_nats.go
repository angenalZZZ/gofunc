package main

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"runtime"
)

func ProdNatS() {
	var (
		name = "cache.>"
		addr = nats.DefaultURL
		ops  = make([]nats.Option, 0)
		nc   *nats.Conn
		err  error
	)

	// "*" matches any token, at any level of the subject.
	// ">" matches any length of the tail of a subject, and can only be the last token
	// E.g. 'cache.>' will match 'cache.set.123', 'cache.get.123', 'cache.del.123'
	if *flagName != "" {
		name = *flagName + ".>"
	}
	if *flagPort > 0 && *flagPort != nats.DefaultPort {
		addr = fmt.Sprintf("nats://127.0.0.1:%d", *flagPort)
	}
	if *flagToken != "" {
		ops = append(ops, nats.Token(*flagToken))
	}
	// If the server requires client certificate
	// E.g. /certs/client-cert.pem  /certs/client-key.pem
	if *flagCert != "" && *flagKey != "" {
		cert := nats.ClientCert(*flagCert, *flagKey)
		ops = append(ops, cert)
	}
	// If you are using a self-signed certificate, you need to have a tls.Config with RootCAs setup
	// E.g. /certs/ca.pem
	if *flagCert != "" {
		cert := nats.RootCAs(*flagCert)
		ops = append(ops, cert)
	}

	if nc, err = nats.Connect(addr, ops...); err != nil {
		_ = fmt.Errorf("Nats failed connect to server: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Nats client connected to %s with token: %s\n", addr, *flagToken)
	defer func() {
		// Drain connection (Preferred for responders)
		// Close() not needed if this is called.
		_ = nc.Drain()
		// Close connection
		//nc.Close()
	}()

	// Requests
	// msg, err := nc.Request("cache.set.123", []byte("456"), time.Second)
	// Replies
	_, err = nc.Subscribe(name, func(m *nats.Msg) {
		result := defaultService.Handle(m.Data)
		if err = m.Respond(result); err != nil {
			_ = fmt.Errorf("Nats failed to Write: %v\n", err)
		}
		//if err = nc.Publish(m.Reply, result); err != nil {
		//	_ = fmt.Errorf("Nats failed to Write: %v\n", err)
		//}
	})
	if err != nil {
		_ = fmt.Errorf("Nats failed to Read: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Nats client subscribed to %s\n", name)

	// Flush connection to server, returns when all messages have been processed.
	_ = nc.Flush()

	// Wait os.Exit
	runtime.Goexit()
}
