package f

import (
	"fmt"
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

// BytesToString convert []byte type to string type.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes convert string type to []byte type.
// NOTE: panic if modify the member value of the []byte.
func StringToBytes(s string) []byte {
	sp := *(*[2]uintptr)(unsafe.Pointer(&s))
	bp := [3]uintptr{sp[0], sp[1], sp[1]}
	return *(*[]byte)(unsafe.Pointer(&bp))
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
