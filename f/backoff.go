package f

import (
	"context"
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-resty/resty/v2"
	"time"
)

var (
	// 重试多次后无法获得结果
	ErrRetryTimesFailure = errors.New("unable to get results after multiple retries")
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

// NewRetryConditionRequest create a retry http request.
// https://github.com/go-resty/resty#retries
func NewRetryConditionRequest(condition func(*resty.Response, error) bool, settings ...func(client *resty.Client)) *resty.Request {
	if condition == nil {
		condition = func(response *resty.Response, err error) bool { return false }
	}
	client := resty.New()
	if len(settings) > 0 {
		settings[0](client)
	}
	return client.AddRetryCondition(condition).R()
}

// NewRetryTimesRequest create a retry http request.
// https://github.com/go-resty/resty#retries
func NewRetryTimesRequest(maxRetries int, waitTime, maxWaitTime time.Duration, settings ...func(client *resty.Client)) *resty.Request {
	if maxRetries < 1 {
		maxRetries = 1
	}
	if waitTime < time.Millisecond {
		waitTime = 100 * time.Millisecond
	}
	if maxWaitTime < time.Second {
		waitTime = 2 * time.Second
	}
	client := resty.New()
	if len(settings) > 0 {
		settings[0](client)
	}
	return client.SetRetryCount(maxRetries).SetRetryWaitTime(waitTime).SetRetryMaxWaitTime(maxWaitTime).R()
	//return resty.Backoff(func() (*resty.Response, error) {
	//	return nil, nil
	//}, resty.Retries(maxRetries), resty.WaitTime(waitTime), resty.MaxWaitTime(maxWaitTime))
}
