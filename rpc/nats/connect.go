package nats

import (
	"github.com/angenalZZZ/gofunc/log"
	"github.com/nats-io/nats.go"
	"time"
)

// New Client Connect.
func New(name, flagAddr, flagCred, flagToken string, flagCert, flagKey string) (nc *nats.Conn, err error) {
	var (
		addr = nats.DefaultURL
		ops  = []nats.Option{nats.Name(name)}
	)

	if Log == nil {
		Log = log.InitConsole("15:04:05.000", false)
	}

	if flagAddr != "" {
		addr = flagAddr
	}
	if flagCred != "" {
		ops = append(ops, nats.UserCredentials(flagCred))
	}
	if flagToken != "" {
		ops = append(ops, nats.Token(flagToken))
	}

	// If the server requires client certificate
	// E.g. /certs/client-cert.pem  /certs/client-key.pem
	if flagCert != "" && flagKey != "" {
		cert := nats.ClientCert(flagCert, flagKey)
		ops = append(ops, cert)
	}
	// If you are using a self-signed certificate, you need to have a tls.Config with RootCAs setup
	// E.g. /certs/ca.pem
	if flagCert != "" {
		cert := nats.RootCAs(flagCert)
		ops = append(ops, cert)
	}

	ops = append(ops,
		nats.MaxReconnects(120),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			Log.Error().Msgf("[nats] disconnected due to: %s, will attempt reconnects for 1m", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			Log.Error().Msgf("[nats] reconnected %q", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			Log.Error().Msgf("[nats] exiting")
		}),
	)

	nc, err = nats.Connect(addr, ops...)
	return
}

// Flush connection to server, returns when all messages have been processed.
func FlushAndCheckLastError(nc *nats.Conn) {
	if err := nc.Flush(); err != nil {
		Log.Error().Msgf("[nats] flush messages error\t>\t%s", err)
	} else if err = nc.LastError(); err != nil {
		Log.Error().Msgf("[nats] after flush and get last error\t>\t%s", err)
	}
}

func SubscribeLimitHandle(sub *nats.Subscription, msgLimit, bytesLimitOfMsg int) {
	if err := sub.SetPendingLimits(msgLimit, msgLimit*bytesLimitOfMsg); err != nil {
		Log.Error().Msgf("[nats] set pending limits error\t>\t%s", err)
	}

	// Delivered returns the number of delivered messages for this subscription.
	if deliveredNum, err := sub.Delivered(); err != nil {
		Log.Error().Msgf("[nats] get number of delivered messages error\t>\t%s", err)
	} else {
		Log.Info().Msgf("[nats] get number of delivered messages: %d \n", deliveredNum)
	}

	// Dropped returns the number of known dropped messages for this subscription.
	if droppedNum, err := sub.Dropped(); err != nil {
		Log.Error().Msgf("[nats] get number of dropped messages error\t>\t%s", err)
	} else {
		Log.Info().Msgf("[nats] get number of dropped messages: %d", droppedNum)
	}
}

func SubscribeErrorHandle(sub *nats.Subscription, async bool, err error) {
	if err != nil {
		Log.Error().Msgf("[nats] failed listening on %q\n %s", sub.Subject, err)
	} else {
		a, v := "async", "valid"
		if async == false {
			a = "sync"
		}
		if sub.IsValid() == false {
			v = "not valid"
		}
		Log.Info().Msgf("[nats] succeeded listening on %s subject %q", v, sub.Subject)
		Log.Info().Msgf("[nats] %s waiting to receive message...", a)
	}
}
