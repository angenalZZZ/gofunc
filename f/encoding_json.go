package f

import (
	json "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const JSON JsonString = ""

type JsonString string

func (j *JsonString) Get(data string) gjson.Result {
	return gjson.Parse(data)
}

func (j *JsonString) Set(data, path string, value interface{}) (string, error) {
	return sjson.Set(data, path, value)
}

// EncodeJson encode a object v to json data.
func EncodeJson(v interface{}) ([]byte, error) {
	return json.ConfigCompatibleWithStandardLibrary.Marshal(v)
}

// DecodeJson decode json data to a object v.
func DecodeJson(data []byte, v interface{}) error {
	return json.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
}
