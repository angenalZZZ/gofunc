///go get github.com/angenalZZZ/gofunc/cmd/jsjob
///go build -ldflags "-s -w" -o A:/test/cmd/jsjob/ ./cmd/jsjob
///start A:/test/cmd/jsjob/jsjob.exe -t
///start A:/test/cmd/jsjob/jsjob.exe -c jsjob.yaml

package main

import (
	"flag"
	"os"
	"runtime"
	"syscall"
	"time"

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

	// Init complete.
	runInit()

	// Run script test.
	runTest()

	// Hot update script file.
	go func() {
		ticket := time.NewTicker(1 * time.Second)
		for range ticket.C {
			isUpdated := false
			for _, job := range jobList {
				if job.FileIsMod() {
					isUpdated = true
					break
				}
			}
			if isUpdated {
				jobCron.Stop()
				jobList = nil
				runInit()
			}
		}
	}()

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		jobCron.Stop()
	})
}
