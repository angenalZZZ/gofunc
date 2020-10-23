package main

import nat "github.com/angenalZZZ/gofunc/rpc/nats"

type handler struct{}

func (hub *handler) Handle(list [nat.BulkSize][]byte) error {
	for _, item := range list {
		if len(item) == 0 {
			break
		}
		if item[0] != '{' {
			nat.Log.Info().Msgf("[nats] received test message on %q: %s", subject, string(item))
		}
	}

	return nil
}
