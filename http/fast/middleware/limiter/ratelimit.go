package limiter

import (
	"errors"
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/angenalZZZ/gofunc/ratelimit"
	"github.com/valyala/fasthttp"
	"time"
)

// ErrTooManyRequests is returned when too many requests.
var ErrTooManyRequests = errors.New("Too many requests, please try again later.")

// RateLimit holds the configuration for the RateLimit middleware handler.
// For example, to allow 100 requests per second and a wait time of 50ms
// to wait for an available token for the request to be allowed, set
// RPS=100, MaxWait=50ms.
type RateLimit struct {
	// RPS is the number of requests per seconds. Tokens will fill at an
	// interval that closely respects that RPS value.
	RPS int64

	// MaxWait is the maximum time to wait for an available token for a
	// request to be allowed. If no token is available, the request is
	// denied without waiting and a status code 429 is returned.
	MaxWait time.Duration

	// Message
	// default: "Too many requests, please try again later."
	Message string

	// StatusCode
	// Default: 429 Too Many Requests
	StatusCode int
}

// NewRateLimiter new RateLimit.
// rps is the number of requests per seconds, msg for the request to be not allowed.
func NewRateLimiter(rps int64, msg string) *RateLimit {
	if rps < 1 {
		rps = 1
	}
	if msg == "" {
		msg = ErrTooManyRequests.Error()
	}
	return &RateLimit{
		RPS:        rps,
		MaxWait:    time.Millisecond, // http request is blocking in milli second
		Message:    msg,
		StatusCode: fasthttp.StatusTooManyRequests,
	}
}

// Wrap returns a handler that allows only the configured number of requests.
// The wrapped handler h is called only if the request is allowed by the rate
// limiter, otherwise a status code 429 is returned.
//
// Each call to Wrap creates a new, distinct rate limiter bucket that controls
// access to h.
func (rl *RateLimit) Wrap(allow func(*fast.Ctx), deny func(*fast.Ctx)) func(*fast.Ctx) {
	rt := ratelimit.NewBucket(time.Second, rl.RPS)
	return func(c *fast.Ctx) {
		if _, ok := rt.TakeMaxDuration(1, rl.MaxWait); ok {
			allow(c) // request to be allowed
			return
		}
		if 0 == rt.Available() {
			rt = ratelimit.NewBucket(time.Second, rl.RPS)
		}
		if deny == nil {
			c.Status(rl.StatusCode).SendString(rl.Message)
		} else {
			deny(c) // request to be deny
		}
	}
}
