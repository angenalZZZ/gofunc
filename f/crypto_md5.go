package f

import (
	"crypto/md5"
	"encoding/hex"
)

func MD5(origData string) string {
	s := md5.New()
	s.Write([]byte(origData))
	return hex.EncodeToString(s.Sum(nil))
}
