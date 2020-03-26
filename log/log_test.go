package log

import (
	"github.com/angenalZZZ/gofunc/configfile"
	"testing"
)

func TestYamlFileConfig(t *testing.T) {
	// 配置选项
	type AppConfig struct {
		Log Config
	}

	// 配置文件
	filename := "log.yaml"
	appConfig := new(AppConfig)
	if err := configfile.YamlTo(filename, appConfig); err != nil {
		t.Fatal(err)
	}

	// 初始化配置
	Log = Init(appConfig.Log)

	// 写日志文件
	Log.Printf("Yaml File: %s", filename)
	Log.Printf("File Config: %#v", appConfig.Log)
	Log.Info().Msg("Test finish.")
}
