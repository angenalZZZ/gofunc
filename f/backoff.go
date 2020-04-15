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
	initialInterval, maxInterval, maxElapsedTime time.Duration, randomizationFactor, multiplier float64,
	ctx ...context.Context) *RetryOperation {
	var o backoff.Operation
	if operation != nil {
		o = operation
	} else {
		o = func() error { return nil }
	}
	if initialInterval == 0 {
		initialInterval = backoff.DefaultInitialInterval
	}
	if maxInterval == 0 {
		maxInterval = backoff.DefaultMaxInterval
	}
	if maxElapsedTime == 0 {
		maxElapsedTime = backoff.DefaultMaxElapsedTime
	}
	if randomizationFactor == 0 {
		randomizationFactor = backoff.DefaultRandomizationFactor
	}
	if multiplier == 0 {
		multiplier = backoff.DefaultMultiplier
	}
	b := &backoff.ExponentialBackOff{
		InitialInterval:     initialInterval,
		RandomizationFactor: randomizationFactor,
		Multiplier:          multiplier,
		MaxInterval:         maxInterval,
		MaxElapsedTime:      maxElapsedTime,
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

func (ro *RetryOperation) NextBackOff() time.Duration {
	select {
	case <-ro.ctx.Done():
		return backoff.Stop
	default:
	}
	next := ro.ExponentialBackOff.NextBackOff()
	if deadline, ok := ro.ctx.Deadline(); ok && deadline.Sub(time.Now()) < next {
		return backoff.Stop
	}
	return next
}

// Retry the operation o until it does not return error or BackOff stops.
func (ro *RetryOperation) Retry() error {
	return backoff.Retry(ro.Operation, ro)
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
