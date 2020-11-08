package main

import (
	"os"
	"strings"
	"time"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/go-redis/redis/v7"
)

var (
	subject    string
	scriptFile string
	configInfo *Config
	configFile = "natsql.yaml"
	cacheDir   = data.CurrentDir
	configMod  time.Time
	scriptMod  time.Time
)

// Config The Config Info For natsql.yaml
type Config struct {
	Db struct {
		Type string
		Conn string
	}
	Nats struct {
		nat.Connection
		Subscribe string
		Amount    int
		Bulk      int
		Interval  int
		Script    string
	}
	Redis *redis.Options
	Log   *log.Config
}

func initConfig() error {
	isConfig := configInfo != nil
	if !isConfig {
		configInfo = new(Config)
	}

	if !isConfig && isConfigMod() == false {
		return os.ErrNotExist
	}

	if err := configfile.YamlTo(configFile, configInfo); err != nil {
		return err
	}

	if filename := configInfo.Nats.Script; strings.HasSuffix(filename, ".js") {
		scriptFile = filename

		if !isConfig && isScriptMod() == false {
			return os.ErrNotExist
		}

		if err := doScriptMod(); err != nil {
			return err
		}
	} else {
		configInfo.Nats.Script = strings.TrimSpace(filename)
	}

	data.DbType, data.DbConn = configInfo.Db.Type, configInfo.Db.Conn

	if store.RedisClient == nil && configInfo.Redis != nil && configInfo.Redis.Addr != "" {
		store.RedisClient = redis.NewClient(configInfo.Redis)
	}

	return nil
}

func isConfigMod() bool {
	if configFile == "" {
		return false
	}
	info, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		return false
	}
	if t := info.ModTime(); t.Unix() != configMod.Unix() {
		configMod = t
		return true
	}
	return false
}

func isScriptMod() bool {
	if scriptFile == "" {
		return false
	}
	info, err := os.Stat(scriptFile)
	if os.IsNotExist(err) {
		return false
	}
	if t := info.ModTime(); t.Unix() != scriptMod.Unix() {
		scriptMod = t
		return true
	}
	return false
}

func doScriptMod() error {
	script, err := f.ReadFile(scriptFile)
	if err != nil {
		return err
	}

	configInfo.Nats.Script = strings.TrimSpace(string(script))
	return nil
}
