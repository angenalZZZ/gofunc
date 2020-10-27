package main

import (
	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/log"
)

var (
	configInfo *Config
	configFile = "natsql.yaml"
	cacheDir   = data.CurrentDir
	subject    = ""
)

// Config The Config Info For natsql.yaml
type Config struct {
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
	configInfo = new(Config)
	return configfile.YamlTo(configFile, configInfo)
}
