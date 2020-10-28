package data

import (
	json "github.com/json-iterator/go"
)

// ObjectJSON gets json object map[string]interface{}
func ObjectJSON(buf []byte) (map[string]interface{}, error) {
	var record map[string]interface{}
	err := json.Unmarshal(buf, &record)
	return record, err
}

// ListJSON gets json list []map[string]interface{}
func ListJSON(buf []byte) ([]map[string]interface{}, error) {
	var records []map[string]interface{}
	err := json.Unmarshal(buf, &records)
	return records, err
}
