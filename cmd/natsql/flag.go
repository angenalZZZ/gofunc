package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/js"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	//flagMsgLimit = flag.Int("c", 100000000, "the nats-Limits for pending messages for this subscription")
	//flagBytesLimit = flag.Int("d", 4096, "the nats-Limits for a message's bytes for this subscription")
	flagConfig   = flag.String("c", "natsql.yaml", "sets config file")
	flagCacheDir = flag.String("d", "", "sets cache persist to disk directory")
	flagTest     = flag.String("t", "", "sets json file and run SQL test")
	flagAddr     = flag.String("a", "", "the NatS-Server address")
	flagName     = flag.String("name", "", "the NatS-Subscription name [required]")
	flagToken    = flag.String("token", "", "the NatS-Token auth string [required]")
	flagCred     = flag.String("cred", "", "the NatS-Cred file")
	flagCert     = flag.String("cert", "", "the NatS-TLS cert file")
	flagKey      = flag.String("key", "", "the NatS-TLS key file")
)

var (
	isTest   = false
	jsonFile string
)

// Flag Parse
func initArgs() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

// Init Config
func checkArgs() {
	if *flagConfig != "" {
		configFile = *flagConfig
	}

	if err := initConfig(); err != nil {
		panic(err)
	}

	if *flagAddr != "" {
		configInfo.Nats.Addr = *flagAddr
	}
	if *flagToken != "" {
		configInfo.Nats.Token = *flagToken
	}
	if *flagCred != "" {
		configInfo.Nats.Cred = *flagCred
	}
	if *flagCert != "" {
		configInfo.Nats.Cert = *flagCert
	}
	if *flagKey != "" {
		configInfo.Nats.Key = *flagKey
	}

	jsonFile = *flagTest
	if jsonFile != "" {
		isTest = true
	}
	if isTest {
		if f.PathExists(jsonFile) == false {
			panic(fmt.Errorf("test json file %q not found", jsonFile))
		}
		configInfo.Log.Level = "debug"
	}

	if log.Log == nil {
		log.Log = log.Init(configInfo.Log)
	}
	if nat.Log == nil {
		nat.Log = log.Log
	}
	js.RunLogTimeFormat = configInfo.Log.TimeFormat

	subject = *flagName
	if subject == "" && !isTest {
		panic("the subscription name can't be empty.")
	}

	if *flagCacheDir != "" {
		cacheDir = *flagCacheDir
	}
	if cacheDir != "" && !isTest && f.PathExists(cacheDir) == false {
		panic("the cache disk directory is not found.")
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

	nat.Log.Debug().Msgf("configuration complete")
}

func runTest(hd *handler) {
	// Check Script
	if err := hd.CheckJs(configInfo.Db.Table.Script); err != nil {
		panic(err)
	}

	if !isTest {
		return
	}

	item, err := f.ReadFile(jsonFile)
	if err != nil {
		panic(fmt.Errorf("test json file %q not opened: %s", jsonFile, err.Error()))
	}

	list, err := data.ListData(item)
	if err != nil {
		panic(err)
	}
	if len(list) == 0 {
		panic(fmt.Errorf("test json file can't be empty"))
	}

	nat.Log.Debug().Msgf("test json file: %q %d records\r\n", jsonFile, len(list))

	if subject == "" {
		subject = "test"
	}
	if err = hd.Handle(list); err != nil {
		panic(err)
	}

	nat.Log.Debug().Msg("test finished.")
	os.Exit(0)
}
