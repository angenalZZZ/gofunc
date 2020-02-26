package configfile

import (
	"gopkg.in/ini.v1"
	"strings"
)

// IniTo 转换ini文件成数据对象.
//
// 例如:
//	err := IniTo("file.ini", &obj)
//
func IniTo(filename string, config interface{}) error {
	return ini.MapTo(config, filename)
}

// IniTos 转换ini文件成多个对象.
//
// 例如:
//  err := IniTos("file.ini", "obj1,obj2", &obj1, &obj2)
//
func IniTos(filename string, sections string, configs ...interface{}) error {
	if cfg, err := ini.Load(filename); err != nil {
		return err
	} else {
		s := strings.Split(sections, ",")
		for i, config := range configs {
			if err = cfg.Section(s[i]).MapTo(config); err != nil {
				return err
			}
		}
	}
	return nil
}

// Ini 生成ini文件.
//
// 例如:
//  err := Ini(&obj, "file.ini")
//
func Ini(config interface{}, filename string) error {
	cfg := ini.Empty()
	if err := ini.ReflectFrom(cfg, config); err != nil {
		return err
	} else {
		return cfg.SaveTo(filename)
	}
}
