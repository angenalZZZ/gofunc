package f

import (
	"reflect"
	"strings"
)

// MapKeys get all keys of the given map.
func MapKeys(mp interface{}) (keys []string) {
	rftVal := reflect.ValueOf(mp)
	if rftVal.Type().Kind() == reflect.Ptr {
		rftVal = rftVal.Elem()
	}

	if rftVal.Kind() != reflect.Map {
		return
	}

	for _, key := range rftVal.MapKeys() {
		keys = append(keys, key.String())
	}
	return
}

// MapValues get all values from the given map.
func MapValues(mp interface{}) (values []interface{}) {
	rftTyp := reflect.TypeOf(mp)
	if rftTyp.Kind() == reflect.Ptr {
		rftTyp = rftTyp.Elem()
	}

	if rftTyp.Kind() != reflect.Map {
		return
	}

	rftVal := reflect.ValueOf(mp)
	for _, key := range rftVal.MapKeys() {
		values = append(values, rftVal.MapIndex(key).Interface())
	}
	return
}

// MapValue get value from a map[string]interface{}. eg "top" "top.sub"
func MapValue(key string, mp map[string]interface{}) (val interface{}, ok bool) {
	if val, ok := mp[key]; ok {
		return val, true
	}

	// has sub key? eg. "top.sub"
	if !strings.ContainsRune(key, '.') {
		return nil, false
	}

	keys := strings.Split(key, ".")
	topK := keys[0]

	// find top item data based on top key
	var item interface{}
	if item, ok = mp[topK]; !ok {
		return
	}

	for _, k := range keys[1:] {
		switch tData := item.(type) {
		case map[string]string: // is simple map
			item, ok = tData[k]
			if !ok {
				return
			}
		case map[string]interface{}: // is map(decode from toml/json)
			if item, ok = tData[k]; !ok {
				return
			}
		case map[interface{}]interface{}: // is map(decode from yaml)
			if item, ok = tData[k]; !ok {
				return
			}
		default: // error
			ok = false
			return
		}
	}

	return item, true
}