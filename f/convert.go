package f

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"unsafe"
)

// String converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
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

// ToJSON convert the input to a valid JSON string
func ToJSON(obj interface{}) (string, error) {
	res, err := EncodeJson(obj)
	if err != nil {
		res = []byte("")
	}
	return string(res), err
}

// ToFloat convert the input string to a float, or 0.0 if the input is not a float.
func ToFloat(str string) (float64, error) {
	res, err := strconv.ParseFloat(str, 64)
	if err != nil {
		res = 0.0
	}
	return res, err
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

// BytesRepeat returns a new byte slice consisting of count copies of b.
func BytesRepeat(b byte, count int) []byte {
	nb := make([]byte, count)
	for i := 0; i < count; i++ {
		nb[i] = b
	}
	return nb
}

// BytesFromPtr converts a pointer to a byte slice without memory allocation.
func BytesFromPtr(p uintptr, b []byte, off int64, size int32) int {
	h := reflect.SliceHeader{}
	h.Cap = int(size)
	h.Len = int(size)
	h.Data = p

	bb := *(*[]byte)(unsafe.Pointer(&h))
	return copy(b, bb[off:size])
}

// BytesToPtr converts a byte slice to a pointer without memory allocation.
func BytesToPtr(b []byte, p uintptr, off int64, size int32) int {
	h := reflect.SliceHeader{}
	h.Cap = int(size)
	h.Len = int(size)
	h.Data = p

	bb := *(*[]byte)(unsafe.Pointer(&h))
	return copy(bb[off:], b)
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

// StringInSlice finds needle in a slice of strings.
func StringInSlice(sliceString []string, needle string) bool {
	for _, b := range sliceString {
		if b == needle {
			return true
		}
	}
	return false
}
