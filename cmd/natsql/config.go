package main

import (
	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/log"
)

var (
	configInfo *config
	configFile = "natsql.yaml"
	cacheDir   = data.CurrentDir
	subject    = ""
)

type config struct {
	Db struct {
		Type  string
		Conn  string
		Table struct {
			Bulk   int
			Script string
		}
	}
	Log *log.Config
}

func initConfig() error {
	configInfo = new(config)
	return configfile.YamlTo(configFile, configInfo)
}
