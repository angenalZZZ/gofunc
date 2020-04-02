package limiter

import (
	"errors"
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/angenalZZZ/gofunc/ratelimit"
	"github.com/valyala/fasthttp"
	"strconv"
	"sync/atomic"
	"time"
)

// ErrTooManyRequests is returned when too many requests.
var ErrTooManyRequests = errors.New("Too many requests, please try again later.")

// RateLimit holds the configuration for the RateLimit middleware handler.
type RateLimit struct {
	// RPS is the number of requests per seconds. Tokens will fill at an
	// interval that closely respects that RPS value.
	RPS int64

	// MaxWait is the maximum time to wait for an available token for a
	// request to be allowed. If no token is available, the request is
	// denied without waiting and a status code 429 is returned.
	MaxWait time.Duration

	// X-Rate-Limit response headers.
	RateLimitHeader func(*fast.Ctx, *ratelimit.Bucket, bool)
	// Http Handler when too many requests.
	DenyHandler func(*fast.Ctx, *ratelimit.Bucket)
}

// NewRateLimiter new RateLimit.
// rps is the number of requests per seconds, msg for the request to be not allowed.
func NewRateLimiterPerSecond(rps int) *RateLimit {
	if rps < 1 {
		rps = 1
	}
	return &RateLimit{
		RPS:     int64(rps),       // RPS is the number of requests per seconds.
		MaxWait: time.Millisecond, // http request is blocking in a milli second.
		RateLimitHeader: func(c *fast.Ctx, rt *ratelimit.Bucket, allowed bool) {
			c.SetHeader("X-Rate-Limit-Duration", "1s")
			c.SetHeader("X-Rate-Limit-Limit", strconv.FormatInt(rt.Capacity(), 10))
			c.SetHeader("X-Rate-Limit-Remaining", strconv.FormatInt(rt.Available(), 10))
			if allowed == false {
				c.SetHeader("X-Rate-Limit-Reset", "1s")
			}
		},
		DenyHandler: func(c *fast.Ctx, rt *ratelimit.Bucket) {
			c.Status(fasthttp.StatusTooManyRequests).SendString(ErrTooManyRequests.Error())
		},
	}
}

// Wrap returns a handler that allows only the configured number of requests.
// The wrapped handler h is called only if the request is allowed by the rate
// limiter, otherwise a status code 429 is returned.
func (rl *RateLimit) Wrap(handler func(*fast.Ctx)) func(*fast.Ctx) {
	ri, rt := int64(0), ratelimit.NewBucket(time.Second, rl.RPS)
	go func() {
		for {
			time.Sleep(time.Millisecond)
			if ri != 0 {
				select {
				case <-time.After(time.Second):
					ri, rt = 0, ratelimit.NewBucket(time.Second, rl.RPS)
				}
			}
		}
	}()
	return func(c *fast.Ctx) {
		// request to be allowed
		if _, ok := rt.TakeMaxDuration(1, rl.MaxWait); ok {
			if ri == 0 {
				atomic.AddInt64(&ri, 1)
			}
			rl.RateLimitHeader(c, rt, true)
			handler(c)
			return
		}
		// request to be deny
		if rl.DenyHandler == nil {
			rl.RateLimitHeader(c, rt, false)
			c.Status(fasthttp.StatusTooManyRequests)
			return
		}
		rl.RateLimitHeader(c, rt, false)
		rl.DenyHandler(c, rt)
	}
}
