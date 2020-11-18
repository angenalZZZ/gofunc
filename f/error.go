package f

import (
	"errors"
	"fmt"
	"strings"
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
	ErrBadInput          = errors.New("invalid input for operation")
)

// goroutine.go
// Errors that are used throughout the Tunny API.
var (
	ErrPoolNotRunning = errors.New("the pool is not running")
	ErrJobNotFunc     = errors.New("generic worker not given a func()")
	ErrWorkerClosed   = errors.New("worker was closed")
	ErrJobTimedOut    = errors.New("job request timed out")
)

// Errors is an array of multiple errors and conforms to the error interface.
type Errors []error

// Errors returns itself.
func (es Errors) Errors() []error {
	return es
}

func (es Errors) Error() string {
	var errs []string
	for _, e := range es {
		errs = append(errs, e.Error())
	}
	return strings.Join(errs, ";")
}

// Error encapsulates a name, an error and whether there's a custom error message or not.
type Error struct {
	Name                     string
	Err                      error
	CustomErrorMessageExists bool

	// Validator indicates the name of the validator that failed
	Validator string
	Path      []string
}

func (e Error) Error() string {
	if e.CustomErrorMessageExists {
		return e.Err.Error()
	}

	errName := e.Name
	if len(e.Path) > 0 {
		errName = strings.Join(append(e.Path, e.Name), ".")
	}

	return errName + ": " + e.Err.Error()
}

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
	if len(args) == 0 {
		panic(format)
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}
