package main

import (
	"strings"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/log"
	"github.com/go-redis/redis/v7"
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
			Amount   int
			Bulk     int
			Interval int
			Script   string
		}
	}
	Redis *redis.Options
	Log   *log.Config
}

func initConfig() error {
	configInfo = new(Config)

	if err := configfile.YamlTo(configFile, configInfo); err != nil {
		return err
	}

	if filename := configInfo.Db.Table.Script; strings.HasSuffix(filename, ".js") {
		script, err := f.ReadFile(filename)
		if err != nil {
			return err
		}
		configInfo.Db.Table.Script = strings.TrimSpace(string(script))
	} else {
		configInfo.Db.Table.Script = strings.TrimSpace(filename)
	}

	if configInfo.Redis != nil && configInfo.Redis.Addr != "" {
		store.RedisClient = redis.NewClient(configInfo.Redis)
	}

	return nil
}
