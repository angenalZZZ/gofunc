package configfile

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// YamlTo 转换yaml文件成数据对象.
//
// 例如:
//	err := YamlTo("file.yaml", &obj)
//
func YamlTo(filename string, config interface{}) error {
	if data, err := ioutil.ReadFile(filename); err != nil {
		return err
	} else if err = yaml.Unmarshal(data, config); err != nil {
		return err
	}
	return nil
}

// YamlToMap 转换yaml文件成map对象.
//
// 例如:
//  m, err := YamlToMap("file.yaml")
//
func YamlToMap(filename string) (config map[interface{}]interface{}, err error) {
	config = make(map[interface{}]interface{})
	if data, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err = yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return
}

// Yaml 生成yaml文件.
//
// 例如:
//  data, err := Yaml(&obj)
//
func Yaml(config interface{}) (data []byte, err error) {
	data, err = yaml.Marshal(config)
	return
}
