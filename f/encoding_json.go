package f

import json "github.com/json-iterator/go"

// EncodeJson encode a object v to json data.
func EncodeJson(v interface{}) ([]byte, error) {
	return json.ConfigCompatibleWithStandardLibrary.Marshal(v)
}

// DecodeJson decode json data to a object v.
func DecodeJson(data []byte, v interface{}) error {
	return json.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
}
