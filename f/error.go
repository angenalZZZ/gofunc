package f

import (
	"errors"
	"fmt"
)

var (
	// backoff.go
	// 重试多次后无法获得结果
	ErrRetryOperationFailure = errors.New("unable to get results after multiple retries")

	// validator.go
	ErrConvertFail       = errors.New("convert value is failure")
	ErrBadComparisonType = errors.New("invalid type for operation")
)

// Must not error, or panic.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustBytes bytes length is these numbers, or panic.
func MustBytes(p []byte, n ...int) {
	if n != nil && len(n) > 0 {
		if p == nil {
			panic(errors.New("wrong empty bytes"))
		}
		ok := false
		l := len(p)
		for _, i := range n {
			if i == l {
				ok = true
				break
			}
		}
		if ok == false {
			panic(errors.New("wrong bytes length"))
		}
	}
}

// Panic exit with error.
func Panic(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}
