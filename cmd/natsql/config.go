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
	// 全局订阅前缀:subject
	subject string
	// 发布订阅功能处理集
	subscribers []*handler
	scriptFile  string
	configInfo  *Config
	configFile  = "natsql.yaml"
	cacheDir    = data.CurrentDir
	configMod   time.Time
	scriptMod   time.Time
)

// Config The Config Info For natsql.yaml
type Config struct {
	// 数据库client
	Db struct {
		// 支持mssql,mysql
		Type string
		// 连接字符串
		Conn string
	}
	// 消息中间件/发布订阅处理
	Nats struct {
		// 消息中间件client
		nat.Connection
		// 全局订阅前缀=功能配置根目录cache+js目录 function func(records)
		Subscribe string
		// 批量获取记录数限制
		Amount int
		// 批量插入记录数<=2000
		Bulk int
		// 间隔?毫秒,批量处理一次
		Interval int
		// 配置订阅任务"subscribe"
		Script string
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
	if scriptFile == "" {
		return nil
	}
	script, err := f.ReadFile(scriptFile)
	if err != nil {
		return err
	}

	configInfo.Nats.Script = strings.TrimSpace(string(script))
	return nil
}
