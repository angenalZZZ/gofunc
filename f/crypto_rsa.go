package f

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// RSAPublicKeyEncrypt 公钥加密.
type RSAPublicKeyEncrypt struct {
	*rsa.PublicKey
}

// NewRSAPublicKeyEncrypt get a RSA Public Key Encrypt.
func NewRSAPublicKeyEncrypt(publicKeyPemBytes []byte) *RSAPublicKeyEncrypt {
	block, _ := pem.Decode(publicKeyPemBytes)
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	Must(err)
	key := publicKey.(*rsa.PublicKey)
	return &RSAPublicKeyEncrypt{PublicKey: key}
}

// EncryptPKCS1v15 encrypts the given message with RSA and the padding
// scheme from PKCS#1 v1.5.  The message must be no longer than the
// length of the public modulus minus 11 bytes.
func (e *RSAPublicKeyEncrypt) EncryptPKCS1v15(origData []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, e.PublicKey, origData)
}

// RSAPrivateKeyDecrypt 私钥解密.
type RSAPrivateKeyDecrypt struct {
	*rsa.PrivateKey
}

// NewRSAPrivateKeyDecrypt get a RSA Private Key Decrypt.
func NewRSAPrivateKeyDecrypt(privateKeyPemBytes []byte) *RSAPrivateKeyDecrypt {
	block, _ := pem.Decode(privateKeyPemBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	Must(err)
	return &RSAPrivateKeyDecrypt{PrivateKey: privateKey}
}

// DecryptPKCS1v15 decrypts a plaintext using RSA and the padding scheme from PKCS#1 v1.5.
// If rand != nil, it uses RSA blinding to avoid timing side-channel attacks.
func (e *RSAPrivateKeyDecrypt) DecryptPKCS1v15(encrypted []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, e.PrivateKey, encrypted)
}
