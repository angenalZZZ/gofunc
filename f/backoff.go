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
	ErrRetryOperationFailure = errors.New("unable to get results after multiple retries")
)

// RetryOperation retry operation manager.
type RetryOperation struct {
	backoff.Operation
	*backoff.ExponentialBackOff
	ctx context.Context
}

// NewRetryOperation create a retryOperation with the context.
func NewRetryOperation(operation func() error,
	InitialInterval, MaxInterval, MaxElapsedTime time.Duration,
	RandomizationFactor, Multiplier float64,
	ctx ...context.Context) *RetryOperation {
	var o backoff.Operation
	if operation != nil {
		o = operation
	} else {
		o = func() error { return nil }
	}
	if InitialInterval == 0 {
		InitialInterval = backoff.DefaultInitialInterval
	}
	if MaxInterval == 0 {
		MaxInterval = backoff.DefaultMaxInterval
	}
	if MaxElapsedTime == 0 {
		MaxElapsedTime = backoff.DefaultMaxElapsedTime
	}
	if RandomizationFactor == 0 {
		RandomizationFactor = backoff.DefaultRandomizationFactor
	}
	if Multiplier == 0 {
		Multiplier = backoff.DefaultMultiplier
	}
	b := &backoff.ExponentialBackOff{
		InitialInterval:     InitialInterval,
		RandomizationFactor: RandomizationFactor,
		Multiplier:          Multiplier,
		MaxInterval:         MaxInterval,
		MaxElapsedTime:      MaxElapsedTime,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	var c context.Context
	if len(ctx) == 1 {
		c = ctx[0]
	} else {
		c = context.Background()
	}
	b.Reset()
	return &RetryOperation{o, b, c}
}

func (ro *RetryOperation) Context() context.Context {
	return ro.ctx
}

// Retry the operation o until it does not return error or BackOff stops.
func (ro *RetryOperation) Retry() error {
	return backoff.Retry(ro.Operation, ro.ExponentialBackOff)
}

// RetryTicker returns a new Ticker containing a channel that will send
// the time at times specified by the BackOff argument. Ticker is
// guaranteed to tick at least once.  The channel is closed when Stop
// method is called or BackOff stops. It is not safe to manipulate the
// provided backoff policy (notably calling NextBackOff or Reset)
// while the ticker is running.
func (ro *RetryOperation) RetryTicker() (err error) {
	ticker := backoff.NewTicker(ro.ExponentialBackOff)
	// Ticks will continue to arrive when the previous operation is still running,
	// so operations that take a while to fail could run in quick succession.
	for range ticker.C {
		if err = ro.Operation(); err != nil {
			continue // will retry...
		}
		ticker.Stop()
		break
	}
	return
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
	for _, setting := range settings {
		setting(client)
	}
	return client.SetRetryCount(maxRetries).SetRetryWaitTime(waitTime).SetRetryMaxWaitTime(maxWaitTime).R()
}
