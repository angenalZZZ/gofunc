package recover

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/angenalZZZ/gofunc/log"
)

// Config defines the config for recover middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Handler is called when a panic occurs
	// Optional. Default: c.SendStatus(500)
	Handler func(*fast.Ctx, error)
	// Log all errors to output
	// Optional. Default: false
	Log bool
	// Output is a writer where logs are written
	// Default: log.Log
	Output log.Logger
}

// New middleware.
// cfg := recover.Config{
//     Log: true,
//     Handler: func(c *fast.Ctx, err error) {
//         c.SendString(err.Error())
//         c.SendStatus(500)
//     },
// }
// app.Use(recover.New(cfg))
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	// Set config if provided
	if len(config) > 0 {
		cfg = config[0]
	}
	// Set config default values
	if cfg.Handler == nil {
		cfg.Handler = func(c *fast.Ctx, err error) {
			c.SendString("unknown error")
			c.SendStatus(500)
		}
	}
	if cfg.Output == nil {
		cfg.Output = log.Log
	}
	// Return middleware handle
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				if cfg.Log && cfg.Output != nil {
					cfg.Output.Err(err).Send()
				}
				cfg.Handler(c, err)
			}
		}()
		c.Next()
	}
}
