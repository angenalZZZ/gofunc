package f

import (
	"errors"
	"fmt"
)

// backoff.go
// Cannot get results after retrying many times
var (
	ErrRetryOperationFailure = errors.New("unable to get results after multiple retries")
)

// validator.go
var (
	ErrConvertFail       = errors.New("convert value is failure")
	ErrBadComparisonType = errors.New("invalid type for operation")
)

// goroutine.go
// Errors that are used throughout the Tunny API.
var (
	ErrPoolNotRunning = errors.New("the pool is not running")
	ErrJobNotFunc     = errors.New("generic worker not given a func()")
	ErrWorkerClosed   = errors.New("worker was closed")
	ErrJobTimedOut    = errors.New("job request timed out")
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
