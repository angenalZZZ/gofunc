package nats

import (
	"time"

	"github.com/angenalZZZ/gofunc/log"
	"github.com/nats-io/nats.go"
)

// Conn represents a bare connection to global nats-server.
// The connection is safe to use in multiple Go routines concurrently.
var Conn *nats.Conn

// Subject global subscription. This can be different
// than the received subject inside a Msg if this is a wildcard.
var Subject string

// Log logger for global Conn.
var Log *log.Logger

// Connection addr and token, credentials, cert and key
// to be used when connecting to a nats-server.
type Connection struct {
	Addr  string
	Token string
	Cred  string
	Cert  string
	Key   string
}

// New creates a client connect.
func New(name, flagAddr, flagCred string, flagToken string, flagCert, flagKey string) (nc *nats.Conn, err error) {
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
	// E.g. /certs/client-cert.pem  /certs/client-Index.pem
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
		nats.MaxReconnects(1200),
		nats.PingInterval(time.Minute),
		nats.ReconnectWait(2*time.Second),
		nats.PingInterval(time.Minute),
		nats.Timeout(2*time.Second),
		nats.SyncQueueLen(100000000),     // sets number of messages will buffer internally.
		nats.ReconnectBufSize(104857600), // 100Mb size of messages kept while busy reconnecting.
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			Log.Error().Msgf("[nats] disconnected due to: %s, will attempt reconnects for 10 minutes", err)
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

// FlushAndCheckLastError Flush connection to server, returns when all messages have been processed.
func FlushAndCheckLastError(nc *nats.Conn) {
	if err := nc.Flush(); err != nil {
		Log.Error().Msgf("[nats] flush messages > %s", err)
	} else if err = nc.LastError(); err != nil {
		Log.Error().Msgf("[nats] after flush and get last error > %s", err)
	}
}

// SubscribeLimitHandle Set pending limits error handle.
func SubscribeLimitHandle(sub *nats.Subscription, msgLimit, bytesLimitOfMsg int) {
	if err := sub.SetPendingLimits(msgLimit, msgLimit*bytesLimitOfMsg); err != nil {
		Log.Error().Msgf("[nats] set pending limits > %s", err)
	}

	// Delivered returns the number of delivered messages for this subscription.
	if deliveredNum, err := sub.Delivered(); err != nil {
		Log.Error().Msgf("[nats] number of messages deliver > %s", err)
	} else if deliveredNum > 0 {
		Log.Info().Msgf("[nats] number of messages deliver: %d", deliveredNum)
	}

	// Dropped returns the number of known dropped messages for this subscription.
	if droppedNum, err := sub.Dropped(); err != nil {
		Log.Error().Msgf("[nats] number of messages dropped > %s", err)
	} else if droppedNum > 0 {
		Log.Info().Msgf("[nats] number of messages dropped: %d", droppedNum)
	}
}

// SubscribeErrorHandle Set listening error handle.
func SubscribeErrorHandle(sub *nats.Subscription, async bool, err error) {
	if err != nil {
		Log.Error().Msgf("[nats] failed listening on %q > %s", sub.Subject, err)
	} else {
		a, v := "async", "available"
		if async == false {
			a = "sync"
		}
		if sub.IsValid() == false {
			v = "invalid"
		}
		Log.Info().Msgf("[nats] successful listening on %s subject: %q", v, sub.Subject)
		Log.Info().Msgf("[nats] start %s waiting to receive messages on %q", a, sub.Subject)
	}
}
