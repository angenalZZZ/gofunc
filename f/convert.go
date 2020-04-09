package f

import (
	"bytes"
	"fmt"
	"github.com/klauspost/crc32"
	"io"
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

// IntSliceRepeat create times.
func IntSliceRepeat(times int, value int) []int {
	q := make([]int, times)
	for i, _ := range q {
		q[i] = value
	}
	return q
}

// IntSliceRepeatAppend append times.
func IntSliceRepeatAppend(slice []int, times int, value int) {
	if slice == nil {
		slice = make([]int, 0, times)
	}
	q := IntSliceRepeat(times, value)
	slice = append(slice, q...)
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

// ToIntCrc32IEEE returns the CRC-32 checksum of data using the IEEE polynomial.
var ToIntCrc32IEEE = crc32.ChecksumIEEE

// String converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringInSlice finds needle in a slice of strings.
func StringInSlice(sliceString []string, needle string) bool {
	for _, b := range sliceString {
		if b == needle {
			return true
		}
	}
	return false
}

// ToString convert number to string
func ToString(val interface{}) (str string) {
	switch tVal := val.(type) {
	case int:
		str = strconv.FormatInt(int64(tVal), 10)
	case int8:
		str = strconv.FormatInt(int64(tVal), 10)
	case int16:
		str = strconv.FormatInt(int64(tVal), 10)
	case int32:
		str = strconv.FormatInt(int64(tVal), 10)
	case int64:
		str = strconv.FormatInt(tVal, 10)
	case uint:
		str = strconv.FormatUint(uint64(tVal), 10)
	case uint8:
		str = strconv.FormatUint(uint64(tVal), 10)
	case uint16:
		str = strconv.FormatUint(uint64(tVal), 10)
	case uint32:
		str = strconv.FormatUint(uint64(tVal), 10)
	case uint64:
		str = strconv.FormatUint(tVal, 10)
	case float32, float64:
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
		} else if t, ok := tVal.(io.Reader); ok {
			buf := new(bytes.Buffer)
			if n, _ := buf.ReadFrom(t); n > 0 {
				return buf.String()
			}
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
