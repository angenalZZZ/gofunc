package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagCont   = flag.Int("c", 1, "total threads")
	flagData   = flag.Int("d", 128, "every time request bytes")
	flagTimes  = flag.Int("t", 1000000, "total times")
	flagRemove = flag.Bool("r", true, "delete data files")
	flagProd   = flag.Bool("prod", false, "run production mode")
	flagAddr   = flag.String("a", "", "the server address")
	flagPort   = flag.Int("p", 6060, "the server port")
	flagSvc    = flag.Int("s", 0, "the server IO transmission mode SHM(SharedMemory)/gRPC/TCP/WS(WebSocket)/NatS")
	flagName   = flag.String("name", "cache", "the Cache name")
	flagToken  = flag.String("token", "", "the Token auth string")
	flagTls    = flag.Bool("tls", false, "connection uses TLS if true, else plain TCP")
	flagCert   = flag.String("cert", "", "the TLS cert file")
	flagKey    = flag.String("key", "", "the TLS key file")
)

func init() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}
