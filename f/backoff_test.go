package f

import (
	"testing"
	"time"
)

func TestNewRetryOperation(t *testing.T) {
	var no int
	ro := NewRetryOperation(func() error {
		if no <= 100 {
			no++
			println(Now().LocalTimeString())
			return ErrRetryOperationFailure
		}
		return nil
	}, time.Millisecond, time.Hour, 0, 0.5, 2.0)
	if err := ro.Retry(); err != nil && err != ErrRetryOperationFailure {
		t.Fatal(err)
	}
}
