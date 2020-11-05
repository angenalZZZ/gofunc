package main

import (
	"os"
	"time"

	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/data/cache/store"
	"github.com/angenalZZZ/gofunc/js"
	"github.com/angenalZZZ/gofunc/log"
	nat "github.com/angenalZZZ/gofunc/rpc/nats"
	"github.com/go-redis/redis/v7"
)

var (
	configInfo *Config
	configFile = "jsjob.yaml"
	configMod  time.Time
)

// Config The Config Info For natsql.yaml
type Config struct {
	Db struct {
		Type string
		Conn string
	}
	Cron  []*js.JobJs
	Nats  *nat.ConnToken
	Redis *redis.Options
	Log   *log.Config
}

func initConfig() error {
	isConfig := configInfo != nil
	if !isConfig {
		configInfo = new(Config)
	}

	if 1 == configMod.Year() && isConfigMod() == false {
		return os.ErrNotExist
	}

	if err := configfile.YamlTo(configFile, configInfo); err != nil {
		return err
	}

	for _, job := range configInfo.Cron {
		if isConfig {
			continue
		}
		if err := job.Init(); err != nil {
			return err
		}
	}

	if store.RedisClient == nil && configInfo.Redis != nil && configInfo.Redis.Addr != "" {
		store.RedisClient = redis.NewClient(configInfo.Redis)
	}

	isConfig = true

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
