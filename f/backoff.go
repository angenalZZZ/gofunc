package f

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/go-resty/resty/v2"
	"time"
)

// RetryOperation retry operation manager.
type RetryOperation struct {
	backoff.Operation
	*backoff.ExponentialBackOff
	ctx context.Context
}

var (
	// Call Periods: 1s,2s,4s,8s,16s,32s,1m,1m,1m... 24h=1day
	RetryPeriodOf1s1m1d = &RetryPeriodDuration{
		InitialInterval: time.Second,
		MaxInterval:     time.Minute,
		MaxElapsedTime:  24 * time.Hour,
		RandomFactor:    0.02,
		Multiplier:      2.0,
	}
)

// RetryPeriodDuration With Periods.
type RetryPeriodDuration struct {
	InitialInterval             time.Duration
	MaxInterval, MaxElapsedTime time.Duration
	RandomFactor, Multiplier    float64
}

// NewRetryOperationWithPeriod create a retryOperation with the context.
func NewRetryOperationWithPeriod(operation func() error, period *RetryPeriodDuration, ctx ...context.Context) *RetryOperation {
	return NewRetryOperation(operation, period.InitialInterval, period.MaxInterval, period.MaxElapsedTime, period.RandomFactor, period.Multiplier, ctx...)
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
