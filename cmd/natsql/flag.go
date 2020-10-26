package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	//flagMsgLimit = flag.Int("c", 100000000, "sets the limits for pending messages for this subscription")
	//flagBytesLimit = flag.Int("d", 4096, "sets the limits for a message's bytes for this subscription")
	flagConfig   = flag.String("c", "natsql.yaml", "sets config file")
	flagCacheDir = flag.String("d", "", "sets cache persist to disk directory")
	flagAddr     = flag.String("a", "", "the server address")
	flagName     = flag.String("name", "", "the subscription name [required]")
	flagToken    = flag.String("token", "", "the Token auth string [required]")
	flagCred     = flag.String("cred", "", "the Cred file")
	flagCert     = flag.String("cert", "", "the TLS cert file")
	flagKey      = flag.String("key", "", "the TLS key file")
)

func init() {
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
}
