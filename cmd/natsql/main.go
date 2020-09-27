///go get github.com/angenalZZZ/gofunc/cmd/natsql
///go build -ldflags "-s -w" -o A:/test/ .

package natsql

import (
	"context"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"os"
	"time"
)

func main() {
	// natsql input args.
	subject, cacheDir := *flagName, data.CurrentDir

	// New Nat's Client Connect.
	nc, err := nat.New("natsql", *flagAddr, "", *flagToken, *flagCert, *flagKey)

	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}

	ctx, wait := f.ContextWithWait(context.Background())

	// Create a subscriber for Client Connect.
	sub := nat.NewSubscriberFastCache(nc, subject, cacheDir)
	sub.MsgLimit = *flagMsgLimit
	sub.BytesLimit = *flagBytesLimit
	sub.Hand = func(list [nat.BulkSize][]byte) error {
		for _, item := range list {
			if len(item) == 0 {
				break
			}
			if item[0] != '{' {
				nat.Log.Error().Msgf("[nats] received test message on %q: %s", sub.Subj, string(item))
			}
		}

		f.DoneContext(ctx)
		nat.Log.Error().Msgf("[nats] test finished on %q", sub.Subj)
		return nil
	}

	// Ping a message.
	go func() {
		time.Sleep(time.Millisecond)
		err = nc.Publish(sub.Subj, []byte("ping"))
		if err != nil {
			nat.Log.Error().Msgf("[nats] failed publishing a test message\t>\t%s", err.Error())
		} else {
			nat.Log.Error().Msgf("[nats] successful publishing a test message")
		}
	}()

	sub.Run(wait)
}
