package log

import (
	"github.com/angenalZZZ/gofunc/configfile"
	. "github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestYamlFileConfig(t *testing.T) {
	// 配置选项
	type AppConfig struct {
		Log PassLagerCfg
	}

	// 配置文件
	filename := "log.yaml"
	appConfig := new(AppConfig)
	if err := configfile.YamlTo(filename, appConfig); err != nil {
		t.Fatal(err)
	}

	// 初始化配置
	Must(InitWithConfig(&appConfig.Log))

	// 写日志文件
	Infof("Yaml File: %s", filename)
	Infof("File Config: %#v", appConfig.Log)
	Infof("Test finish.")
}
