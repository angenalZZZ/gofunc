package f

import (
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto/z"
	"github.com/klauspost/crc32"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Int32Sum returns the CRC-32 checksum of data using the IEEE polynomial.
var Int32Sum = crc32.ChecksumIEEE

// Int32SumString returns the CRC-32 checksum of data using the IEEE polynomial.
var Int32SumString = func(s string) uint32 {
	var b []byte
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
	bh.Len = len(s)
	bh.Cap = len(s)
	return crc32.ChecksumIEEE(b)
}

// Int64Sum xx.hash is a Go implementation of the 64-bit xxHash algorithm, XXH64.
var Int64Sum = xxhash.Sum64

// Int64SumString xx.hash is a Go implementation of the 64-bit xxHash algorithm, XXH64.
var Int64SumString = xxhash.Sum64String

// Int64MemHash is the hash function used by go map,
// it utilizes available hardware instructions(behaves as aes.hash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
var Int64MemHash = z.MemHash

// Int64MemHashString is the hash function used by go map,
// it utilizes available hardware instructions (behaves as aes.hash if aes instruction is available).
// NOTE: The hash seed changes for every process. So, this cannot be used as a persistent hash.
var Int64MemHashString = z.MemHashString

// Int convert string to int64, or return 0.
func Int(v interface{}) (i int64) {
	i, _ = ToInt(v, false)
	return
}

// ToInt parse string to int64, or return ErrConvertFail.
func ToInt(v interface{}, strict ...bool) (i int64, err error) {
	switch notParse := len(strict) > 0 && strict[0] == true; t := v.(type) {
	case string:
		if notParse {
			return 0, ErrConvertFail
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
		if notParse {
			return 0, ErrConvertFail
		}
		i = int64(t)
	case float64:
		if notParse {
			return 0, ErrConvertFail
		}
		i = int64(t)
	default:
		err = ErrConvertFail
	}
	return
}
