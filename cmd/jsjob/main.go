///go get github.com/angenalZZZ/gofunc/cmd/jsjob
///go build -ldflags "-s -w" -o A:/test/ ./cmd/jsjob
///start A:/test/jsjob.exe -t
///start A:/test/jsjob.exe -c jsjob.yaml

package main

import (
	"flag"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/robfig/cron/v3"
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

	// New DB Connect.
	data.DbType, data.DbConn = configInfo.Db.Type, configInfo.Db.Conn

	// New Client Connect.
	nat.Subject = "jsjob"
	nat.Conn, err = nat.New("jsjob", configInfo.Nats.Addr, "", configInfo.Nats.Token, "", "")
	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}

	// Run script test.
	runTest()

	// Run script job.
	logger := &log.CronLogger{Log: log.Log}
	jobs := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(logger),
	))
	for _, job := range configInfo.Cron {
		// Adds a Job to the Cron
		if _, err := jobs.AddJob(job.Spec, job); err != nil {
			log.Log.Error().Msgf("[jsjob] failed add %q to cron: %v\n", job.Name, err)
			os.Exit(1)
		}
	}
	jobs.Start()

	// Hot update script file.
	go func() {
		ticket := time.NewTicker(1 * time.Second)
		for range ticket.C {
			for _, job := range configInfo.Cron {
				if !job.FileIsMod() {
					continue
				}
				if err := job.FileMod(); err == nil {
					log.Log.Info().Msgf("Hot update script %q file.", job.Name)
				} else {
					log.Log.Info().Msgf("Hot update script %q file error: %v", job.Name, err)
				}
			}
		}
	}()

	// Pass the signals you want to end your application.
	death := f.NewDeath(syscall.SIGINT, syscall.SIGTERM)
	// When you want to block for shutdown signals.
	death.WaitForDeathWithFunc(func() {
		jobs.Stop()
	})
}
