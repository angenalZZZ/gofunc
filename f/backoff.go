package f

import (
	"context"
	"github.com/cenkalti/backoff/v4"
)

// Retry the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once.
// If o returns a *PermanentError, the operation is not retried, and the
// wrapped error is returned.
// Retry sleeps the goroutine for the duration returned by BackOff after a
// failed operation returns.
//
// Creates an instance of ExponentialBackOff using default values.
//
// operation must not be nil
func Retry(operation func() error) error {
	return backoff.Retry(operation, backoff.NewExponentialBackOff())
}

// RetryContext the operation o until it does not return error or BackOff stops.
// o is guaranteed to be run at least once. a BackOffContext with context ctx.
//
// operation must not be nil
// ctx must not be nil, or equals context.Background()
func RetryContext(operation func() error, ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.Retry(operation, b)
}

// RetryTicker returns a new Ticker containing a channel that will send
// the time at times specified by the BackOff argument. Ticker is
// guaranteed to tick at least once.  The channel is closed when Stop
// method is called or BackOff stops. It is not safe to manipulate the
// provided backoff policy (notably calling NextBackOff or Reset)
// while the ticker is running.
//
// operation must not be nil
func RetryTicker(operation func() error) (err error) {
	ticker := backoff.NewTicker(backoff.NewExponentialBackOff())

	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = operation(); err != nil {
			// will retry...
			continue
		}

		ticker.Stop()
		break
	}
	return err
}
