package f

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

// CryptoMD5 md5 hash.
func CryptoMD5(origData string) string {
	s := md5.New()
	s.Write([]byte(origData))
	return hex.EncodeToString(s.Sum(nil))
}

// CryptoHmac hmac SHA256|SHA384|SHA512 hash.
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA256, base64.URLEncoding.EncodeToString)
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA384, EncodeBase64RawURL)
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA512, EncodeBase64URL)
func CryptoHmac(origData, key string, hash crypto.Hash, encode func(src []byte) string) string {
	h := hmac.New(hash.New, []byte(key))
	_, err := h.Write([]byte(origData))
	if err == nil {
		if encode == nil {
			encode = EncodeBase64RawURL
		}
		return encode(h.Sum(nil))
	}
	return ""
}
