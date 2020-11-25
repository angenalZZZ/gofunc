///go get github.com/angenalZZZ/gofunc/cmd/jsrun
///go build -ldflags "-s -w" -o A:/test/cmd/jsrun/ ./cmd/jsrun
///start A:/test/cmd/jsrun/jsrun.exe jsrun.js

package main

import (
	"flag"
	"os"
	"runtime"
	"syscall"

	"github.com/angenalZZZ/gofunc/f"
)

func main() {
	// Your Arguments.
	initArgs()
	if len(os.Args) < 2 {
		flag.Usage()
		return
	}

	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(runtime.NumCPU()))

	// Check Arguments And Init Config.
	checkArgs()

	// New Client Connect.
	natClientConnect()

	// Run script.
	run()

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {})
}
