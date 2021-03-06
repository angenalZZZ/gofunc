///go get github.com/angenalZZZ/gofunc/cmd/natsql
///go build -ldflags "-s -w" -o A:/test/cmd/natsql/natsql.exe ./cmd/natsql
///start A:/test/cmd/natsql/natsql.exe -t data.json
///start A:/test/cmd/natsql/natsql.exe -c natsql.yaml

// CSharp(4CPU+16G+MHD).benchmark PUB:60K~100Kq/s SUB:100K~1Mq/s

package main

import (
	"flag"
	"os"
	"runtime"
	"syscall"
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

	// Init complete.
	runInit()

	// Run script test.
	runTest()

	// Hot update script file.
	go func() {
		ticket := time.NewTicker(2 * time.Second)
		for range ticket.C {
			if !isScriptMod() {
				continue
			}
			if err := doScriptMod(); err == nil {
				createHandlers() // Init Subscribers And Handlers.
				nat.Log.Info().Msg("Hot update natsql.yaml")
			} else {
				nat.Log.Info().Msg("Hot update natsql.yaml error: " + err.Error())
			}
		}
	}()
	go func() {
		ticket := time.NewTicker(2 * time.Second)
		for range ticket.C {
			for _, h := range handlers {
				if !h.isScriptMod() {
					continue
				}
				_ = h.doScriptMod()
				nat.Log.Info().Msgf("Hot update %q natsql.js", h.sub.Subj)
			}
		}
	}()

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		stopHandlers() // Stop Subscribers And Handlers.
	})
}
