package f

import (
	json "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type Json string

// Parse parses the json and returns a result.
//
// This function expects that the json is well-formed, and does not validate.
// Invalid json will not panic, but it may return back unexpected results.
// If you are consuming JSON from an unpredictable source then you may want to
// use the Valid function first.
func (j Json) Parse() gjson.Result {
	return gjson.Parse(string(j))
}

// Get searches json for the specified path.
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

// Set sets a json value for the specified path.
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

// String gets string from Json.
func (j Json) String() string {
	return string(j)
}

// EncodeJson encode a object v to json data.
func EncodeJson(v interface{}) ([]byte, error) {
	return json.ConfigCompatibleWithStandardLibrary.Marshal(v)
}

// DecodeJson decode json data to a object v.
func DecodeJson(data []byte, v interface{}) error {
	return json.ConfigCompatibleWithStandardLibrary.Unmarshal(data, v)
}
