package f

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
)

// NewServerTLSConfig Setup a bare-bones TLS config for the server.
func NewServerTLSConfig(nextProto string) *tls.Config {
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

// NewClientTLSConfig Setup a bare-bones TLS config for the client.
func NewClientTLSConfig(nextProto string) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{nextProto},
	}
}
