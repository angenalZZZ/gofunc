package f

import (
	"context"
	"testing"
)

func TestNewRetryOperation(t *testing.T) {
	var no int
	ctx, cancel := context.WithCancel(context.Background())

	// Call Periods: 1s,2s,4s,8s,16s,32s,1m,1m,1m... 24h=1day
	ro := NewRetryOperationWithPeriod(func() error {
		if no <= 100 {
			no++
			println(Now().LocalTimeString())
			if no == 10 {
				cancel()
			}
			return ErrRetryOperationFailure
		}
		return nil
	}, RetryPeriodOf1s1m1d, ctx)

	// Call Retry Start.
	if err := ro.Retry(); err != nil && err != ErrRetryOperationFailure {
		t.Fatal(err)
	}
}
