package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/angenalZZZ/gofunc/js"

	"github.com/robfig/cron/v3"

	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
)

var (
	flagConfig = flag.String("c", "jsjob.yaml", "sets config file")
	flagTest   = flag.Bool("t", false, "run test")
	flagAddr   = flag.String("a", "", "the NatS-Server address")
	flagToken  = flag.String("token", "", "the NatS-Token auth string [required]")
	flagCred   = flag.String("cred", "", "the NatS-Cred file")
	flagCert   = flag.String("cert", "", "the NatS-TLS cert file")
	flagKey    = flag.String("key", "", "the NatS-TLS key file")
)

var (
	jobCron *cron.Cron
	jobList []*js.JobJs
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

	if log.Log == nil {
		log.Log = log.Init(configInfo.Log)
	}
	if nat.Log == nil {
		nat.Log = log.Log
	}

	log.Log.Debug().Msgf("configuration complete")
}

func clientConnect() {
	var err error

	// NatS
	nat.Subject = "jsjob"
	nat.Conn, err = nat.New("jsjob", configInfo.Nats.Addr, configInfo.Nats.Cred, configInfo.Nats.Token, configInfo.Nats.Cert, configInfo.Nats.Key)
	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}
}

func runInit() {
	if *flagTest == true {
		return
	}

	var (
		r   = getRuntime()
		err error
	)
	defer func() { r.ClearInterrupt() }()

	// load js jobs
	jobList, err = js.NewJobs(r, scriptFile, "cron", "")
	if err != nil {
		log.Log.Error().Msgf("[jsjob] %v\n", err)
		os.Exit(1)
	}

	// init jobCron
	logger := &log.CronLogger{Log: log.Log}
	jobCron = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(logger),
	))

	// adds jobs to the cron
	for _, job := range jobList {
		job.R = getRuntime
		if _, err = jobCron.AddJob(job.Spec, job); err != nil {
			log.Log.Error().Msgf("[jsjob] failed add %q to cron: %v\n", job.Name, err)
			os.Exit(1)
		}
	}

	jobCron.Start()
}

func runTest() {
	if *flagTest == false {
		return
	}

	var (
		r   = getRuntime()
		err error
	)
	defer func() { r.ClearInterrupt() }()

	// load js jobs
	jobList, err = js.NewJobs(r, scriptFile, "cron", "")
	if err != nil {
		log.Log.Error().Msgf("[test] %v\n", err)
		os.Exit(1)
	}

	// adds jobs to the cron
	for _, job := range jobList {
		job.R = getRuntime
		job.Run()
	}

	log.Log.Debug().Msg("test finished.")
	os.Exit(0)
}
