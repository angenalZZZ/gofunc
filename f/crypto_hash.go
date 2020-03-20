package f

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// CryptoMD5 md5 hash.
func CryptoMD5(origData string) string {
	s := md5.New()
	s.Write([]byte(origData))
	return hex.EncodeToString(s.Sum(nil))
}

// CryptoHS256 hmac sha256 hash.
func CryptoHS256(origData, privateKey string, encode func(src []byte) string) string {
	hash := hmac.New(sha256.New, []byte(privateKey))
	_, err := hash.Write([]byte(origData))
	if err == nil {
		if encode == nil {
			encode = EncodeBase64RawURL
		}
		return encode(hash.Sum(nil))
	}
	return ""
}
