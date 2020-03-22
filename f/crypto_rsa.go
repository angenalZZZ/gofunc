package f

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// RSAPublicKey 公钥加密或验签.
type RSAPublicKey struct {
	*rsa.PublicKey
}

// NewRSAPublicKey get a RSA Public Key Encrypt.
func NewRSAPublicKey(publicKeyPemBytes []byte) *RSAPublicKey {
	block, _ := pem.Decode(publicKeyPemBytes)
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	Must(err)
	key := publicKey.(*rsa.PublicKey)
	return &RSAPublicKey{PublicKey: key}
}

// EncryptPKCS1v15 encrypts the given message with RSA and the padding
// scheme from PKCS#1 v1.5.  The message must be no longer than the
// length of the public modulus minus 11 bytes.
func (e *RSAPublicKey) EncryptPKCS1v15(origData []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, e.PublicKey, origData)
}

// VerifyPKCS1v15 verifies an RSA PKCS#1 v1.5 signature.
// err := VerifyPKCS1v15(origData, sig, crypto.SHA256)
// err := VerifyPKCS1v15(origData, sig, crypto.SHA384)
// err := VerifyPKCS1v15(origData, sig, crypto.SHA512)
func (e *RSAPublicKey) VerifyPKCS1v15(origData, sig []byte, hash crypto.Hash) error {
	hasher := hash.New()
	hasher.Write(origData)
	return rsa.VerifyPKCS1v15(e.PublicKey, hash, hasher.Sum(nil), sig)
}

// RSAPrivateKey 私钥解密或签名.
type RSAPrivateKey struct {
	*rsa.PrivateKey
}

// NewRSAPrivateKey get a RSA Private Key Decrypt.
func NewRSAPrivateKey(privateKeyPemBytes []byte) *RSAPrivateKey {
	block, _ := pem.Decode(privateKeyPemBytes)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	Must(err)
	return &RSAPrivateKey{PrivateKey: privateKey}
}

// DecryptPKCS1v15 decrypts a plaintext using RSA and the padding scheme from PKCS#1 v1.5.
// If rand != nil, it uses RSA blinding to avoid timing side-channel attacks.
func (e *RSAPrivateKey) DecryptPKCS1v15(encrypted []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, e.PrivateKey, encrypted)
}

// SignPKCS1v15 calculates the signature of hashed using RSA-PKCS1-V1_5-SIGN from RSA PKCS#1 v1.5.
// sig, err := SignPKCS1v15(origData, crypto.SHA256)
// sig, err := SignPKCS1v15(origData, crypto.SHA384)
// sig, err := SignPKCS1v15(origData, crypto.SHA512)
func (e *RSAPrivateKey) SignPKCS1v15(origData []byte, hash crypto.Hash) ([]byte, error) {
	hasher := hash.New()
	hasher.Write(origData)
	return rsa.SignPKCS1v15(rand.Reader, e.PrivateKey, hash, hasher.Sum(nil))
}
