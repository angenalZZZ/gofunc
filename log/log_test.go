package log

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/configfile"
	"testing"
)

func TestYamlFileConfig(t *testing.T) {
	// 配置文件
	filename := "log.yaml"
	logCfg := new(AConfig)
	if err := configfile.YamlTo(filename, logCfg); err != nil {
		t.Fatal(err)
	}

	// 初始化配置
	Log = Init(logCfg.Log)

	// 写日志
	Log.Debug().Msgf("Yaml File: %s, %d:%d", filename)
	Log.Info().Str("Config", fmt.Sprintf("%+v", logCfg.Log)).Send()
	Log.Info().Timestamp().Msg("Test finish.\n ok!")
}
