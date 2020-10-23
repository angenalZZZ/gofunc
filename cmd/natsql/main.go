///go get github.com/angenalZZZ/gofunc/cmd/natsql
///go build -ldflags "-s -w" -o A:/test/ ./cmd/natsql
///start A:/test/natsql.exe -name Test -token HGJ766GR767FKJU0

package main

import (
	"flag"
	"os"
	"runtime"

	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

func main() {
	// Your Arguments.
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	// Check Arguments And Init Config.
	checkArgs()
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	// New Client Connect.
	nc, err := nat.New(subject, *flagAddr, *flagCred, *flagToken, *flagCert, *flagKey)

	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}

	// Create a subscriber for Client Connect.
	sub, hd := nat.NewSubscriberFastCache(nc, subject, cacheDir), new(handler)
	//sub.MsgLimit = *flagMsgLimit
	//sub.BytesLimit = *flagBytesLimit
	sub.Hand = hd.Handle

	// Ping a message.
	//go func() {
	//	time.Sleep(time.Millisecond)
	//	err = nc.Publish(sub.Subj, []byte("ping"))
	//	if err != nil {
	//		nat.Log.Error().Msgf("[nats] failed publishing a test message\t>\t%s", err.Error())
	//	} else {
	//		nat.Log.Info().Msgf("[nats] successful publishing a test message")
	//	}
	//}()

	// Waiting to exit.
	sub.Run()
}
