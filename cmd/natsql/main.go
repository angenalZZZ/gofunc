///go get github.com/angenalZZZ/gofunc/cmd/natsql
///go build -ldflags "-s -w" -o A:/test/ ./cmd/natsql
///start A:/test/natsql.exe -t data.json
///start A:/test/natsql.exe -name Test -token HGJ766GR767FKJU0

package main

import (
	"flag"
	"os"
	"runtime"
	"time"

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
	var err error
	// New Client Connect.
	nat.Subject = subject
	nat.Conn, err = nat.New(subject, *flagAddr, *flagCred, *flagToken, *flagCert, *flagKey)
	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}

	// Create global js objects.
	jsObj := make(map[string]interface{})
	jsObj["config"] = configInfo

	// Run script test.
	hd := new(handler)
	hd.jsObj = jsObj
	runTest(hd)

	// Create global subscriber for Client Connect.
	sub := nat.NewSubscriberFastCache(nat.Conn, subject, cacheDir)
	sub.Hand = hd.Handle
	sub.LimitAmount(int64(configInfo.Db.Table.Amount), time.Duration(configInfo.Db.Table.Interval)*time.Millisecond)
	//sub.LimitMessage(*flagMsgLimit, *flagBytesLimit)
	dump := "NatS Config Info:\r\n {\"Subj\":%q,\"CacheDir\":%q,\"MsgLimit\":%d,\"BytesLimit\":%d,\"OnceAmount\":%d,\"OnceInterval\":%s}"
	nat.Log.Debug().Msgf(dump, sub.Subj, sub.CacheDir, sub.MsgLimit, sub.BytesLimit, sub.OnceAmount, sub.OnceInterval)

	// Waiting to exit.
	sub.Run()
}
