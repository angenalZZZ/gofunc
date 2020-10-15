package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	//flagMsgLimit = flag.Int("c", 100000000, "sets the limits for pending messages for this subscription")
	//flagBytesLimit = flag.Int("d", 4096, "sets the limits for a message's bytes for this subscription")
	flagCacheDir = flag.String("d", "", "sets cache persist to disk directory")
	flagAddr     = flag.String("a", "", "the server address")
	flagName     = flag.String("name", "", "the subscription name [required]")
	flagToken    = flag.String("token", "", "the Token auth string [required]")
	flagCert     = flag.String("cert", "", "the TLS cert file")
	flagKey      = flag.String("key", "", "the TLS key file")
)

func init() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}
