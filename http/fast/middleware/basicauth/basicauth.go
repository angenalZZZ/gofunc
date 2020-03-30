package basicauth

import (
	"encoding/base64"
	"github.com/angenalZZZ/gofunc/http/fast"
	"strings"
)

// Config defines the config for BasicAuth middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Users defines the allowed credentials
	// Required. Default: map[string]string{}
	Users map[string]string
	// Realm is a string to define realm attribute of BasicAuth.
	// the realm identifies the system to authenticate against
	// and can be used by clients to save credentials
	// Optional. Default: "Restricted".
	Realm string
	// Authorize defines a function you can pass
	// to check the credentials however you want.
	// It will be called with a username and password
	// and is expected to return true or false to indicate
	// that the credentials were approved or not.
	// Optional. Default: nil.
	Authorize func(string, string) bool
	// Unauthorized defines the response body for unauthorized responses.
	// Optional. Default: nil
	Unauthorized func(*fast.Ctx)
}

// New middleware.
//  cfg := basicauth.Config{
//    Users: map[string]string{
//      "admin":  "123456",
//    },
//  }
//  app.Use(basicauth.New(cfg))
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Users == nil {
		cfg.Users = map[string]string{}
	}
	if cfg.Realm == "" {
		cfg.Realm = "Restricted"
	}
	if cfg.Authorize == nil {
		cfg.Authorize = func(user, pass string) bool {
			if user == "" || pass == "" {
				return false
			}
			if _pass, ok := cfg.Users[user]; ok {
				return _pass == pass
			}
			return false
		}
	}
	if cfg.Unauthorized == nil {
		cfg.Unauthorized = func(c *fast.Ctx) {
			c.SetHeader("WWW-Authenticate", "basic realm="+cfg.Realm)
			c.SendStatus(401)
		}
	}
	// Return middleware handler
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		// GetHeader authorization header
		auth := c.GetHeader("Authorization")
		// Check if header is valid
		if len(auth) > 6 && strings.ToLower(auth[:5]) == "basic" {
			// Try to decode
			if raw, err := base64.StdEncoding.DecodeString(auth[6:]); err == nil {
				// Convert to string
				cred := string(raw)
				// Find semi column
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Split into user & pass
						user := cred[:i]
						pass := cred[i+1:]
						// If exist & match in Users, we let him pass
						if cfg.Authorize(user, pass) {
							c.Next()
							return
						}
					}
				}
			}
		}
		// Authentication failed
		cfg.Unauthorized(c)
	}
}
