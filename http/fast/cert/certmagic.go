package cert

import (
	"crypto/tls"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/caddyserver/certmagic"
)

// CertMagicConfig to do.
// https://github.com/caddyserver/certmagic
func CertMagicConfig(certEmail string, domains ...string) *tls.Config {
	// provide an email address
	certmagic.DefaultACME.Email = certEmail
	// use the staging endpoint while we're developing
	certmagic.DefaultACME.CA = certmagic.LetsEncryptStagingCA

	cfg, err := certmagic.TLS(domains)
	f.Must(err)
	return cfg
}
