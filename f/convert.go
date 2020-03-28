package f

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Bool convert string to bool
func Bool(s string) bool {
	ok, _ := ToBool(s)
	return ok
}

// ToBool parse string to bool
func ToBool(s string) (bool, error) {
	lower := strings.ToLower(s)
	switch lower {
	case "1", "on", "yes", "true":
		return true, nil
	case "0", "off", "no", "false":
		return false, nil
	}
	return false, fmt.Errorf("'%s' cannot convert to bool", s)
}

// Int convert string to int64
func Int(v interface{}) (i int64) {
	i, _ = ToInt(v, false)
	return
}

// ToInt parse string to int64
func ToInt(v interface{}, strict bool) (i int64, err error) {
	switch t := v.(type) {
	case string:
		if strict {
			return 0, errConvertFail
		}
		i, err = strconv.ParseInt(strings.TrimSpace(t), 10, 0)
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		i = t
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		i = int64(t)
	case float32:
	case float64:
		if strict {
			return 0, errConvertFail
		}
		i = int64(t)
	default:
		err = errConvertFail
	}
	return
}

// String converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToString convert number to string
func ToString(val interface{}) (str string) {
	switch tVal := val.(type) {
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
		str = strconv.FormatInt(tVal, 10)
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		str = strconv.FormatUint(tVal, 10)
	case float32:
	case float64:
		str = fmt.Sprintf("%g", tVal)
	case string:
		str = tVal
	case []byte:
		str = String(tVal)
	case nil:
		str = ""
	default:
		if t, ok := tVal.(fmt.Stringer); ok {
			str = t.String()
		} else {
			str = fmt.Sprintf("%v", tVal)
		}
	}
	return
}

// Bytes converts string to a byte slice without memory allocation.
// NOTE: panic if modify the member value of the []byte.
func Bytes(s string) (b []byte) {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// ToBytes converts string to a byte slice without memory allocation.
// NOTE: panic if modify the member value of the []byte.
func ToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{Data: sh.Data, Len: sh.Len, Cap: sh.Len}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// StringsConvert converts the string slice to a new slice using fn.
// If fn returns error, exit the conversion and return the error.
func StringsConvert(a []string, fn func(string) (string, error)) ([]string, error) {
	ret := make([]string, len(a))
	for i, s := range a {
		r, err := fn(s)
		if err != nil {
			return nil, err
		}
		ret[i] = r
	}
	return ret, nil
}

// StringsConvertMap converts the string slice to a new map using fn.
// If fn returns error, exit the conversion and return the error.
func StringsConvertMap(a []string, fn func(string) (string, error)) (map[string]string, error) {
	ret := make(map[string]string, len(a))
	for _, s := range a {
		r, err := fn(s)
		if err != nil {
			return nil, err
		}
		ret[s] = r
	}
	return ret, nil
}
