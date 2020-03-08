package f

import (
	"crypto/md5"
	"encoding/hex"
)

// CryptoMD5 md5 hash.
func CryptoMD5(origData string) string {
	s := md5.New()
	s.Write([]byte(origData))
	return hex.EncodeToString(s.Sum(nil))
}
