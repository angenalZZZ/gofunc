package f

import (
	"fmt"
	"strconv"
	"strings"
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
		i = int64(t)
	case int8:
		i = int64(t)
	case int16:
		i = int64(t)
	case int32:
		i = int64(t)
	case int64:
		i = t
	case uint:
		i = int64(t)
	case uint8:
		i = int64(t)
	case uint16:
		i = int64(t)
	case uint32:
		i = int64(t)
	case uint64:
		i = int64(t)
	case float32:
		if strict {
			return 0, errConvertFail
		}
		i = int64(t)
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
		str = fmt.Sprint(tVal)
	case float64:
		str = fmt.Sprint(tVal)
	case string:
		str = tVal
	case nil:
		str = ""
	default:
		str = fmt.Sprintf("%+v", tVal)
	}
	return
}
