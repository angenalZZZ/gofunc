package limiter

import (
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/angenalZZZ/gofunc/ratelimit"
	"github.com/valyala/fasthttp"
	"time"
)

// RateLimit holds the configuration for the RateLimit middleware handler.
// For example, to allow 100 requests per second and a wait time of 50ms
// to wait for an available token for the request to be allowed, set
// RPS=100, MaxWait=50ms.
type RateLimit struct {
	// RPS is the number of requests per seconds. Tokens will fill at an
	// interval that closely respects that RPS value.
	RPS int64

	// Capacity is the maximum number of tokens that can be available in
	// the bucket. The bucket starts at full capacity. If the capacity is
	// <= 0, it is set to the RPS.
	Capacity int64

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
		msg = "Too many requests, please try again later."
	}
	return &RateLimit{
		RPS:        rps,
		Capacity:   rps,
		MaxWait:    10 * time.Millisecond, // http request is blocking in 10 milli second
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
func (rl *RateLimit) Wrap(h func(*fast.Ctx)) func(*fast.Ctx) {
	bucket := ratelimit.NewBucketWithRate(float64(rl.RPS), rl.Capacity)
	return func(c *fast.Ctx) {
		if bucket.WaitMaxDuration(1, rl.MaxWait) {
			h(c) // request to be allowed
			return
		}
		c.Status(rl.StatusCode).SendString(rl.Message)
	}
}
