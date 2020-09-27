package natsql

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagMsgLimit   = flag.Int("c", 10000000, "sets the limits for pending messages for this subscription")
	flagBytesLimit = flag.Int("d", 4096, "sets the limits for a message's bytes for this subscription")
	flagAddr       = flag.String("a", "", "the server address")
	flagName       = flag.String("name", "cache", "the subscription name")
	flagToken      = flag.String("token", "", "the Token auth string")
	flagCert       = flag.String("cert", "", "the TLS cert file")
	flagKey        = flag.String("key", "", "the TLS key file")
)

func init() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}
