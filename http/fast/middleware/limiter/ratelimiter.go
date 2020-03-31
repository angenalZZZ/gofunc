package limiter

import (
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

// Config defines the config for limiter middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Timeout in seconds on how long to keep records of requests in memory
	// Default: 1
	Timeout int64
	// Max number of recent connections during `Timeout` seconds before sending a 429 response
	// Default: 100
	Max int64
	// Message
	// default: "Too many requests, please try again later."
	Message string
	// StatusCode
	// Default: 429 Too Many Requests
	StatusCode int
	// Key allows to use a custom handler to create custom keys
	// Default: func(c *fast.Ctx) string {
	//   return c.IP()
	// }
	Key func(*fast.Ctx) string
	// Handler is called when a request hits the limit
	// Default: func(c *fast.Ctx) {
	//   c.Status(cfg.StatusCode).SendString(cfg.Message)
	// }
	Handler func(*fast.Ctx)
}

// New middleware.
//  cfg := limiter.Config{
//    Timeout: 1,
//    Max: 100,
//  }
// app.Use(limiter.New(cfg))
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 1
	}
	if cfg.Max == 0 {
		cfg.Max = 100
	}
	if cfg.Message == "" {
		cfg.Message = "Too many requests, please try again later."
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = fasthttp.StatusTooManyRequests
	}
	if cfg.Key == nil {
		cfg.Key = func(c *fast.Ctx) string {
			return c.IP()
		}
	}
	if cfg.Handler == nil {
		cfg.Handler = func(c *fast.Ctx) {
			c.Status(cfg.StatusCode).SendString(cfg.Message)
		}
	}
	// Limiter settings
	var hits = f.NewConcurrentMap()
	var reset = f.NewConcurrentMap()
	var timestamp = time.Now().UnixNano()
	// Update timestamp every second
	go func() {
		for {
			timestamp = time.Now().UnixNano()
			time.Sleep(time.Microsecond)
		}
	}()
	// Reset hits every cfg.Timeout
	go func() {
		for {
			// For every key in reset
			for item := range reset.IterBuffered() {
				// If resetTime exist and current time is equal or bigger
				i := item.Val.(int64)
				if i != 0 && timestamp >= i {
					// Reset hits and resetTime
					hits.Set(item.Key, 0)
					reset.Set(item.Key, 0)
				}
			}
			// Wait cfg.Timeout
			time.Sleep(time.Duration(cfg.Timeout) * time.Second)
		}
	}()
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		// GetHeader key (default is the remote IP)
		key := cfg.Key(c)
		// Increment key hits
		var hitCount int64
		if hit, ok := hits.Get(key); ok {
			hitCount = hit.(int64) + 1
		} else {
			hitCount = 1
		}
		hits.Set(key, hitCount)
		// SetHeader unix timestamp if not exist
		var hitReset int64
		if hit, ok := reset.Get(key); ok {
			hitReset = hit.(int64)
		} else {
			hitReset = timestamp + cfg.Timeout
			reset.Set(key, hitReset)
		}
		// SetHeader how many hits we have left
		remaining := cfg.Max - hitCount
		// Calculate when it resets in seconds
		resetTime := hitReset - timestamp
		// Check if hits exceed the cfg.Max
		if remaining < 1 {
			// Call Handler func
			cfg.Handler(c)
			// Return response with Retry-After header
			// https://tools.ietf.org/html/rfc6584
			c.SetHeader("Retry-After", strconv.FormatInt(resetTime, 10))
			return
		}
		// We can continue, update RateLimit headers
		c.SetHeader("X-RateLimit-Limit", strconv.FormatInt(cfg.Max, 10))
		c.SetHeader("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		c.SetHeader("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
		c.Next()
	}
}
