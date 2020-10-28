package data

import (
	"bytes"

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

// ListData gets json list of item RawMessage []byte
func ListData(buf []byte) ([][]byte, error) {
	var s [][]byte
	buf = bytes.TrimSpace(buf)
	if c0 := bytes.IndexByte(buf, '['); c0 > 0 {
		buf = buf[c0:]
	}
	if buf[0] == '[' && buf[len(buf)-1] == ']' {
		var records []map[string]interface{}
		err := json.Unmarshal(buf, &records)
		if err != nil {
			return s, err
		}
		for _, item := range records {
			b, _ := json.Marshal(item)
			s = append(s, b)
		}
	} else {
		s = append(s, buf)
	}
	return s, nil
}
