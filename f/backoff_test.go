package f_test

import (
	"context"
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestNewRetryOperation(t *testing.T) {
	var no int
	ctx, cancel := context.WithCancel(context.Background())

	// Call Periods: 1s,2s,4s,8s,16s,32s,1m,1m,1m... 24h=1day
	ro := f.NewRetryOperationWithPeriod(func() error {
		if no <= 100 {
			no++
			println(f.Now().LocalTimeString())
			if no == 10 {
				cancel()
			}
			return f.ErrRetryOperationFailure
		}
		return nil
	}, f.RetryPeriodOf1s1m1d, ctx)

	// Call Retry Start
	if err := ro.Retry(); err != nil && err != f.ErrRetryOperationFailure {
		t.Fatal(err)
	}
}
