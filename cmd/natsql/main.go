///go get github.com/angenalZZZ/gofunc/cmd/natsql
///go build -ldflags "-s -w" -o A:/test/ ./cmd/natsql
///start A:/test/natsql.exe -name Test -token HGJ766GR767FKJU0

package main

import (
	"flag"
	"os"
	"runtime"
	"time"

	"github.com/angenalZZZ/gofunc/f"

	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

func main() {
	// Your Arguments.
	initArgs()
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	// Check Arguments And Init Config.
	checkArgs()
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))
	nat.Log.Debug().Msgf("NatSql Config Info:\r\n %s", f.EncodedJson(configInfo))

	// New Client Connect.
	nc, err := nat.New(subject, *flagAddr, *flagCred, *flagToken, *flagCert, *flagKey)

	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}

	// Create a subscriber for Client Connect.
	sub, hd := nat.NewSubscriberFastCache(nc, subject, cacheDir), new(handler)
	sub.Hand = hd.Handle
	sub.LimitAmount(int64(configInfo.Db.Table.Bulk), time.Duration(configInfo.Db.Table.Interval)*time.Millisecond)
	//sub.LimitMessage(*flagMsgLimit, *flagBytesLimit)
	nat.Log.Debug().Msgf("NatS Config Info:\r\n {\"Subj\":%q,\"CacheDir\":%q,\"MsgLimit\":%d,\"BytesLimit\":%d,\"OnceAmount\":%d,\"OnceInterval\":%s}",
		sub.Subj, sub.CacheDir, sub.MsgLimit, sub.BytesLimit, sub.OnceAmount, sub.OnceInterval)

	// Waiting to exit.
	sub.Run()
}
