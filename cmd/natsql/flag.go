package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dop251/goja"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	//flagMsgLimit = flag.Int("c", 100000000, "the nats-Limits for pending messages for this subscription")
	//flagBytesLimit = flag.Int("d", 4096, "the nats-Limits for a message's bytes for this subscription")
	flagConfig   = flag.String("c", "natsql.yaml", "sets config file")
	flagCacheDir = flag.String("d", "", "sets cache persist to disk directory")
	flagAddr     = flag.String("a", "", "the NatS-Server address")
	flagName     = flag.String("name", "", "the NatS-Subscription name [required]")
	flagToken    = flag.String("token", "", "the NatS-Token auth string [required]")
	flagCred     = flag.String("cred", "", "the NatS-Cred file")
	flagCert     = flag.String("cert", "", "the NatS-TLS cert file")
	flagKey      = flag.String("key", "", "the NatS-TLS key file")
)

func initArgs() {
	// Flag Parse
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func checkArgs() {
	// Init Config
	if *flagConfig != "" {
		configFile = *flagConfig
	}

	if err := initConfig(); err != nil {
		panic(err)
	}
	nat.Log = log.Init(configInfo.Log)

	subject = *flagName
	if subject == "" {
		panic("the subscription name can't be empty.")
	}

	if *flagCacheDir != "" {
		cacheDir = *flagCacheDir
	}
	if cacheDir != "" && f.PathExists(cacheDir) == false {
		panic("the cache disk directory isn't be exists.")
	}

	// Check Script
	vm := goja.New()
	vm.Set("records", []map[string]interface{}{})
	if _, err := vm.RunString(configInfo.Db.Table.Script); err != nil {
		panic("the table script error, must contain array 'records'.")
	}

	if configInfo.Db.Table.Amount < 1 {
		configInfo.Db.Table.Amount = -1
	}
	if configInfo.Db.Table.Bulk < 1 {
		configInfo.Db.Table.Bulk = 1
	}
	if configInfo.Db.Table.Amount > 0 && configInfo.Db.Table.Amount < configInfo.Db.Table.Bulk {
		configInfo.Db.Table.Amount = configInfo.Db.Table.Bulk
	}
	if configInfo.Db.Table.Interval < 1 {
		configInfo.Db.Table.Interval = 1
	}
}
