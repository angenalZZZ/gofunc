package requestid

import (
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/rs/xid"
)

const HeaderXRequestID string = "X-Request-ID"

// Config defines the config for X-Request-ID middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Generator defines a function to generate an ID.
	// Optional. Default: func() string {
	//   return uuid.New().String()
	// }
	Generator func() string
}

// New middleware.
// app.Use(requestid.New())
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	// SetHeader config if provided
	if len(config) > 0 {
		cfg = config[0]
	}
	// Set config default values
	if cfg.Generator == nil {
		cfg.Generator = func() string {
			return xid.New().String()
		}
	}
	// Return middleware handle
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		// Get value from X-Request-ID
		rid := c.GetHeader(HeaderXRequestID)
		// Create new ID
		if rid == "" {
			rid = cfg.Generator()
			// Set X-Request-ID
			c.SetHeader(HeaderXRequestID, rid)
		}
		c.Next()
	}
}
