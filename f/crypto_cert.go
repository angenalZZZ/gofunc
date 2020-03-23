package f

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net/http"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// NewTLSServerAutoCertConfig serve over tls with autoCerts from let's encrypt.
func NewTLSServerAutoCertConfig(email string, domains string) *tls.Config {
	certDomains := strings.Split(domains, " ")
	certManager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      email,                                  // Email for problems with certs
		HostPolicy: autocert.HostWhitelist(certDomains...), // Domains to request certs for
		Cache:      autocert.DirCache("secrets"),           // Cache certs in secrets folder
	}

	return &tls.Config{
		// Pass in a cert manager if you want one set
		// this will only be used if the server Certificates are empty
		GetCertificate: certManager.GetCertificate,

		// VersionTLS11 or VersionTLS12 would exclude many browsers
		// inc. Android 4.x, IE 10, Opera 12.17, Safari 6
		// So unfortunately not acceptable as a default yet
		// Current default here for clarity
		MinVersion: tls.VersionTLS10,

		// Causes servers to use Go's default cipherSuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
	}
}

// NewTLSServerTestConfig Setup a bare-bones TLS config for the server.
func NewTLSServerTestConfig(nextProto string) *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	Must(err)

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certPEMBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	Must(err)

	keyPEMBytes := x509.MarshalPKCS1PrivateKey(key)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certPEMBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyPEMBytes})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	Must(err)

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{nextProto},
	}
}

// NewTLSClientTestConfig Setup a bare-bones TLS config for the client.
func NewTLSClientTestConfig(nextProto string) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, // 忽略服务器证书校验; 忽略自签名的服务器证书.
		NextProtos:         []string{nextProto},
	}
}

// NewHttpsTransportSkipVerify 用于 Client Dial 忽略服务器证书校验; 忽略自签名的服务器证书.
func NewHttpsTransportSkipVerify() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}
