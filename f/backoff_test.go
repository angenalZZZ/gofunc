package f

import (
	"context"
	"testing"
	"time"
)

func TestNewRetryOperation(t *testing.T) {
	var no int

	// Interval period: 1s,2s,4s,8s,16s,32s,1m,1m,1m... 24h=1day
	initialInterval, maxInterval, maxElapsedTime, randomizationFactor, multiplier :=
		time.Second, time.Minute, 24*time.Hour, 0.02, 2.0

	ctx, cancel := context.WithCancel(context.Background())

	ro := NewRetryOperation(func() error {
		if no <= 100 {
			no++
			println(Now().LocalTimeString())
			if no == 10 {
				cancel()
			}
			return ErrRetryOperationFailure
		}
		return nil
	}, initialInterval, maxInterval, maxElapsedTime, randomizationFactor, multiplier, ctx)
	if err := ro.Retry(); err != nil && err != ErrRetryOperationFailure {
		t.Fatal(err)
	}
}
