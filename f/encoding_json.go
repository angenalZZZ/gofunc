package f

import (
	"github.com/angenalZZZ/gofunc/g"
	json "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"strings"
)

type Json string

// NewJson New Json string.
func NewJson(s string) Json {
	if s == "" {
		return ""
	}
	if j := strings.TrimSpace(s); j != "" && len(j) > 1 {
		return Json(j)
	}
	return ""
}

// Parse parses the json and returns a result.
//
// This function expects that the json is well-formed, and does not validate.
// Invalid json will not panic, but it may return back unexpected results.
// If you are consuming JSON from an unpredictable source then you may want to
// use the Valid function first.
func (j Json) Parse() gjson.Result {
	return gjson.Parse(string(j))
}

// GetHeader searches json for the specified path.
// A path is in dot syntax, such as "name.last" or "age".
// When the value is found it's returned immediately.
//
// A path is a series of keys searated by a dot.
// A key may contain special wildcard characters '*' and '?'.
// To access an array value use the index as the key.
// To get the number of elements in an array or to access a child path, use
// the '#' character.
// The dot and wildcard character can be escaped with '\'.
//
//  {
//    "name": {"first": "Tom", "last": "Anderson"},
//    "age":37,
//    "children": ["Sara","Alex","Jack"],
//    "friends": [
//      {"first": "James", "last": "Murphy"},
//      {"first": "Roger", "last": "Craig"}
//    ]
//  }
//  "name.last"          >> "Anderson"
//  "age"                >> 37
//  "children"           >> ["Sara","Alex","Jack"]
//  "children.#"         >> 3
//  "children.1"         >> "Alex"
//  "child*.2"           >> "Jack"
//  "c?ildren.0"         >> "Sara"
//  "friends.#.first"    >> ["James","Roger"]
//
// This function expects that the json is well-formed, and does not validate.
// Invalid json will not panic, but it may return back unexpected results.
// If you are consuming JSON from an unpredictable source then you may want to
// use the Valid function first.
func (j Json) Get(path string) gjson.Result {
	return gjson.Get(string(j), path)
}

// GetMany searches json for the multiple paths.
// The return value is a Result array where the number of items
// will be equal to the number of input paths.
func (j Json) GetMany(path ...string) []gjson.Result {
	return gjson.GetMany(string(j), path...)
}

// SetHeader sets a json value for the specified path.
// A path is in dot syntax, such as "name.last" or "age".
// This function expects that the json is well-formed, and does not validate.
// Invalid json will not panic, but it may return back unexpected results.
// An error is returned if the path is not valid.
func (j Json) Set(path string, value interface{}) (string, error) {
	return sjson.Set(string(j), path, value)
}

// Sets return a json value.
func (j Json) Sets(path string, value interface{}) Json {
	if s, err := j.Set(path, value); err == nil {
		return Json(s)
	}
	return j
}

// Delete deletes a value from json for the specified path.
func (j Json) Delete(path string) (string, error) {
	return sjson.Delete(string(j), path)
}

// Deletes return a json value.
func (j Json) Deletes(path string) Json {
	if s, err := sjson.Delete(string(j), path); err == nil {
		return Json(s)
	}
	return j
}

// Map unmarshal to a map.
func (j Json) Map() map[string]interface{} {
	if m, ok := gjson.Parse(string(j)).Value().(map[string]interface{}); ok {
		return m
	}
	return nil
}

// Each to ForEachLine will iterate through JSON lines.
func (j Json) Each(iterator func(m map[string]interface{})) {
	gjson.ForEachLine(string(j), func(line gjson.Result) bool {
		if m, ok := line.Value().(map[string]interface{}); ok {
			iterator(m)
		}
		return true
	})
}

// String gets string from Json.
func (j Json) String() string {
	return string(j)
}

// HasValue gets json not equals empty.
func (j Json) HasValue() bool {
	return j != ""
}

// IsValid Check json string.
func (j Json) IsValid() bool {
	return gjson.Valid(string(j))
}

// Exists Check for the existence of a value.
func (j Json) Exists(path string) bool {
	return gjson.Get(string(j), path).Exists()
}

// EncodedJson returns json data.
func EncodedJson(v interface{}) []byte {
	if p, err := json.Marshal(v); err != nil {
		return []byte{}
	} else {
		return p
	}
}

// EncodedMap returns json data.
func EncodedMap(v interface{}) []byte {
	if m1, ok := v.(map[string]interface{}); ok {
		m := make(map[string]interface{})
		for k, o := range m1 {
			if o != nil {
				m[k] = o
			}
		}
		return EncodedJson(m)
	}
	if m2, ok := v.(g.Map); ok {
		m, m1 := make(map[string]interface{}), map[string]interface{}(m2)
		for k, o := range m1 {
			if o != nil {
				m[k] = o
			}
		}
		return EncodedJson(m)
	}
	return EncodedJson(v)
}

// EncodeJson encode a object v to json data.
func EncodeJson(v interface{}) ([]byte, error) {
	return json.ConfigCompatibleWithStandardLibrary.Marshal(v)
}

// DecodeJson decode json data to a object v.
func DecodeJson(data []byte, v interface{}) error {
	return json.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
}
