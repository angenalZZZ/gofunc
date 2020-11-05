package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	flagConfig = flag.String("c", "jsjob.yaml", "sets config file")
	flagTest   = flag.Bool("t", false, "run test")
)

func initArgs() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func checkArgs() {
	if *flagConfig != "" {
		configFile = *flagConfig
	}

	if err := initConfig(); err != nil {
		panic(err)
	}

	if log.Log == nil {
		log.Log = log.Init(configInfo.Log)
	}
	if nat.Log == nil {
		nat.Log = log.Log
	}

	log.Log.Debug().Msgf("configuration complete")
}

func runTest() {
	if *flagTest {
		for _, job := range configInfo.Cron {
			job.Run()
		}

		log.Log.Debug().Msg("test finished.")
		os.Exit(0)
	}
}
