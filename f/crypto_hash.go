package f

import (
	"crypto"
	"crypto/hmac"
	"encoding/hex"
	"fmt"
	"reflect"
)

// CryptoMD5 md5 hash.
func CryptoMD5(origData string) string {
	s := crypto.MD5.New()
	s.Write(Bytes(origData))
	return hex.EncodeToString(s.Sum(nil))
}

// CryptoMD5Key md5 hash key.
func CryptoMD5Key(key interface{}) string {
	digest := crypto.MD5.New()
	_, _ = fmt.Fprint(digest, reflect.TypeOf(key))
	_, _ = fmt.Fprint(digest, key)
	hash := digest.Sum(nil)
	return fmt.Sprintf("%x", hash)
}

// CryptoHmac hmac SHA256|SHA384|SHA512 hash.
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA256, base64.URLEncoding.EncodeToString)
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA384, EncodeBase64RawURL)
// encryptedBytes, err := CryptoHmac(origData, key, crypto.SHA512, EncodeBase64URL)
func CryptoHmac(origData, key string, hash crypto.Hash, encode func(src []byte) string) string {
	h := hmac.New(hash.New, Bytes(key))
	_, err := h.Write(Bytes(origData))
	if err == nil {
		if encode == nil {
			encode = EncodeBase64RawURL
		}
		return encode(h.Sum(nil))
	}
	return ""
}
