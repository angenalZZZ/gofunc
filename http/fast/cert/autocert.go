package cert

import (
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
)

// AutoCertConfig to do.
// Let’s Encrypt has rate limits: https://letsencrypt.org/docs/rate-limits/
// It's recommended to use it's staging environment to test the code:
// https://letsencrypt.org/docs/staging-environment/
func AutoCertConfig(certDir string, domains ...string) *tls.Config {
	// Certificate manager
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		// Replace with your domain
		HostPolicy: autocert.HostWhitelist(domains...),
		// Folder to store the certificates
		Cache: autocert.DirCache(certDir),
	}

	// TLS Config
	return &tls.Config{
		// Get Certificate from Let's Encrypt
		GetCertificate: m.GetCertificate,
		// By default NextProtos contains the "h2"
		// This has to be removed since Fasthttp does not support HTTP/2
		// Or it will cause a flood of PRI method logs
		// http://webconcepts.info/concepts/http-method/PRI
		NextProtos: []string{
			"http/1.1", "acme-tls/1",
		},
	}
}
