package cors

import (
	"github.com/angenalZZZ/gofunc/http/fast"
	"strconv"
	"strings"
)

// Config defines the config for cors middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Optional. Default value []string{"*"}.
	AllowOrigins []string
	// Optional. Default value "GET,POST,HEAD,PUT,DELETE,PATCH"
	AllowMethods string
	// Optional. Default value "".
	AllowHeaders string
	// Optional. Default value false.
	AllowCredentials bool
	// Optional. Default value "".
	ExposeHeaders string
	// Optional. Default value 0.
	MaxAge int64
	// X-XSS-Protection...
	X bool
}

// New middleware.
//  cfg := cors.Config{
//    AllowHeaders: "authorization, origin, content-type, accept",
//    MaxAge: 86400,
//    X: true,
//  }
// app.Use(cors.New(cfg))
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if len(cfg.AllowOrigins) == 0 {
		cfg.AllowOrigins = []string{"*"}
	}
	if cfg.AllowMethods == "" {
		cfg.AllowMethods = "GET,POST,HEAD,PUT,DELETE,PATCH"
	}
	// Middleware function
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		allowOrigin := ""
		// Check allowed origins
		for _, o := range cfg.AllowOrigins {
			if o == "*" && cfg.AllowCredentials {
				allowOrigin = origin
				break
			}
			if o == "*" || o == origin {
				allowOrigin = o
				break
			}
			if matchSubDomain(origin, o) {
				allowOrigin = origin
				break
			}
		}
		// Simple request
		if c.Method() != "OPTIONS" {
			c.Vary("Origin")
			c.SetHeader("Access-Control-Allow-Origin", allowOrigin)

			if cfg.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}
			if cfg.ExposeHeaders != "" {
				c.SetHeader("Access-Control-Expose-Headers", cfg.ExposeHeaders)
			}
			if cfg.X {
				c.XSSProtection()
			}
			c.Next()
			return
		}
		// Preflight request
		c.Vary("Origin")
		c.Vary("Access-Control-Request-Method")
		c.Vary("Access-Control-Request-Headers")
		c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
		c.SetHeader("Access-Control-Allow-Methods", cfg.AllowMethods)

		if cfg.AllowCredentials {
			c.SetHeader("Access-Control-Allow-Credentials", "true")
		}
		if cfg.AllowHeaders != "" {
			c.SetHeader("Access-Control-Allow-Headers", cfg.AllowHeaders)
		} else {
			h := c.GetHeader("Access-Control-Request-Headers")
			if h != "" {
				c.SetHeader("Access-Control-Allow-Headers", h)
			}
		}
		if cfg.MaxAge > 0 {
			c.SetHeader("Access-Control-Max-Age", strconv.FormatInt(cfg.MaxAge, 10))
		}
		if cfg.X {
			c.XSSProtection()
		}
		c.SendStatus(204) // No Content
	}
}

// find domain
func matchScheme(domain, pattern string) bool {
	i := strings.Index(domain, ":")
	p := strings.Index(pattern, ":")
	return i != -1 && p != -1 && domain[:i] == pattern[:p]
}

// compares authority with wildcard
func matchSubDomain(domain, pattern string) bool {
	if !matchScheme(domain, pattern) {
		return false
	}
	i := strings.Index(domain, "://")
	p := strings.Index(pattern, "://")
	if i == -1 || p == -1 {
		return false
	}
	domAuth := domain[i+3:]
	// to avoid long loop by invalid long domain
	if len(domAuth) > 253 {
		return false
	}
	patAuth := pattern[p+3:]
	domComp := strings.Split(domAuth, ".")
	patComp := strings.Split(patAuth, ".")
	for i := len(domComp)/2 - 1; i >= 0; i-- {
		opp := len(domComp) - 1 - i
		domComp[i], domComp[opp] = domComp[opp], domComp[i]
	}
	for i := len(patComp)/2 - 1; i >= 0; i-- {
		opp := len(patComp) - 1 - i
		patComp[i], patComp[opp] = patComp[opp], patComp[i]
	}
	for i, v := range domComp {
		if len(patComp) <= i {
			return false
		}
		p := patComp[i]
		if p == "*" {
			return true
		}
		if p != v {
			return false
		}
	}
	return false
}
