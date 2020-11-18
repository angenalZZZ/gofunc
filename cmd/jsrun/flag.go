package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/js"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	flagConfig = flag.String("c", "jsrun.yaml", "set config file")
	flagDaemon = flag.Bool("d", false, "set as daemons")
)

func initArgs() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func inputArg() (string, error) {
	if terminal.IsTerminal(0) == false {
		buf, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return f.String(buf), nil
	}

	args := flag.Args()
	if len(args) > 2 {
		if args[0] == "-c" {
			return args[2], nil
		}
		return args[0], nil
	}

	return "", f.ErrBadInput
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

	js.RunLogTimeFormat = configInfo.Log.TimeFormat

	filename, err := inputArg()
	if err != nil {
		exit(err)
	}
	if filename != "" {
		scriptFile = filename
	}

	log.Log.Debug().Msgf("configuration complete.")
}

func natClientConnect() {
	var err error

	// NatS
	nat.Subject = "jsjob"
	nat.Conn, err = nat.New("jsjob", configInfo.Nats.Addr, configInfo.Nats.Cred, configInfo.Nats.Token, configInfo.Nats.Cert, configInfo.Nats.Key)
	if err != nil {
		nat.Log.Error().Msgf("[nats] failed connect to server: %v\n", err)
		os.Exit(1)
	}
}

func run() {
	var r = js.NewRuntime(nil)
	defer func() { r.Clear() }()

	// load js
	if strings.HasSuffix(scriptFile, ".js") {
		if isScriptMod() == false {
			exit(os.ErrNotExist)
		}
		if err := doScriptMod(); err != nil {
			exit(err)
		}
	} else {
		configInfo.Script = strings.TrimSpace(scriptFile)
	}

	if configInfo.Script == "" {
		exit(f.ErrBadInput)
	}

	if _, err := r.RunString(configInfo.Script); err != nil {
		exit(err)
	}

	log.Log.Debug().Msg("[js] run finished.")

	if *flagDaemon == false {
		os.Exit(0)
	}
}

func exit(err error) {
	nat.Log.Error().Msgf("[js] run error: %v\n", err)
	os.Exit(0)
}
