package configfile

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// UnmarshalYaml 转换yaml文件成数据对象.
//
// 例如:
//	err := UnmarshalYaml('file.yaml', &obj)
//
func UnmarshalYaml(filename string, config interface{}) error {
	if data, err := ioutil.ReadFile(filename); err != nil {
		return err
	} else if err = yaml.Unmarshal(data, config); err != nil {
		return err
	}
	return nil
}

// UnmarshalYamlToMap 转换yaml文件成map对象.
//
// 例如:
//  m, err := UnmarshalYamlToMap('file.yaml')
//
func UnmarshalYamlToMap(filename string) (config map[interface{}]interface{}, err error) {
	config = make(map[interface{}]interface{})
	if data, err := ioutil.ReadFile(filename); err != nil {
		return config, err
	} else if err = yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}
	return
}

// MarshalYaml 生成yaml文件.
//
// 例如:
//  data, err := MarshalYaml(&obj)
//
func MarshalYaml(config interface{}) (data []byte, err error) {
	data, err = yaml.Marshal(config)
	return
}
