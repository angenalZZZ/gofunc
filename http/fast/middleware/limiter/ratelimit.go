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
	RateLimitHeader func(*fast.Ctx, *RateLimit, bool)
	// Http Handler when too many requests.
	DenyHandler func(*fast.Ctx, *RateLimit)

	// rate limit Bucket
	*ratelimit.Bucket
	start int32
}

// NewRateLimiter new RateLimit.
// rps is the number of requests per seconds, msg for the request to be not allowed.
func NewRateLimiterPerSecond(rps int) *RateLimit {
	capacity := int64(rps)
	if capacity < 1 {
		capacity = 1
	}
	return &RateLimit{
		RPS:     capacity,         // RPS is the number of requests per seconds.
		MaxWait: time.Millisecond, // http request is blocking in a milli second.
		Bucket:  ratelimit.NewBucket(time.Second, capacity),
		RateLimitHeader: func(c *fast.Ctx, r *RateLimit, allowed bool) {
			//c.SetHeader("X-Rate-Limit-Duration", "1s")
			c.SetHeader("X-Rate-Limit-Limit", strconv.FormatInt(r.Bucket.Capacity(), 10))
			c.SetHeader("X-Rate-Limit-Remaining", strconv.FormatInt(r.Bucket.Available(), 10))
			c.SetHeader("X-Rate-Limit-Reset", "1s")
			if allowed == false {
				c.SetHeader("Retry-After", "1s")
			}
		},
		DenyHandler: func(c *fast.Ctx, r *RateLimit) {
			c.Status(fasthttp.StatusTooManyRequests).SendString(ErrTooManyRequests.Error())
		},
	}
}

// Wrap returns a handler that allows only the configured number of requests.
// The wrapped handler h is called only if the request is allowed by the rate
// limiter, otherwise a status code 429 is returned.
func (rl *RateLimit) Wrap(handler func(*fast.Ctx)) func(*fast.Ctx) {
	go rl.Refill()
	return func(c *fast.Ctx) {
		// request to be allowed
		if _, ok := rl.Bucket.TakeMaxDuration(1, rl.MaxWait); ok {
			if rl.start == 0 {
				atomic.AddInt32(&rl.start, 1)
			}
			rl.RateLimitHeader(c, rl, true)
			handler(c)
			return
		}
		// request to be deny
		if rl.DenyHandler == nil {
			rl.RateLimitHeader(c, rl, false)
			c.Status(fasthttp.StatusTooManyRequests)
			return
		}
		rl.RateLimitHeader(c, rl, false)
		rl.DenyHandler(c, rl)
	}
}

// Renew refill, start new rate limit Bucket.
func (rl *RateLimit) Refill() {
	a := time.Second - rl.MaxWait
	for {
		time.Sleep(time.Microsecond)
		if rl.start == 0 {
			continue
		}
		select {
		case <-time.After(a):
			rl.Restart()
		}
	}
}

// Restart start new rate limit Bucket.
func (rl *RateLimit) Restart() {
	rl.start, rl.Bucket = 0, ratelimit.NewBucket(time.Second, rl.RPS)
}
