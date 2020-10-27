package main

import (
	"io/ioutil"
	"strings"

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

	if err := configfile.YamlTo(configFile, configInfo); err != nil {
		return err
	}

	if filename := configInfo.Db.Table.Script; strings.HasSuffix(filename, ".js") {
		script, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		configInfo.Db.Table.Script = strings.TrimSpace(string(script))
	} else {
		configInfo.Db.Table.Script = strings.TrimSpace(filename)
	}

	return nil
}
