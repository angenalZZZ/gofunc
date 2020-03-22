package f

import (
	"encoding/base64"
	"strings"
)

// EncodeBase64Trim Base64 Trim= encoded string.
// encodedString := EncodeBase64Trim(srcBytes, EncodeBase64URL)
func EncodeBase64Trim(src []byte, encode func(src []byte) string) string {
	return strings.TrimRight(encode(src), "=")
}

// DecodeBase64Trim Base64 Trim= decoded bytes.
// srcBytes, err := DecodeBase64Trim(encodedString, DecodeBase64URL)
func DecodeBase64Trim(s string, decode func(s string) ([]byte, error)) ([]byte, error) {
	if l := len(s) % 4; l > 0 {
		s += strings.Repeat("=", 4-l)
	}
	return decode(s)
}

// EncodeBase64Std base64.StdEncoding.EncodeToString
func EncodeBase64Std(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// EncodeBase64URL base64.URLEncoding.EncodeToString
func EncodeBase64URL(src []byte) string {
	return base64.URLEncoding.EncodeToString(src)
}

// EncodeBase64RawURL base64.RawURLEncoding.EncodeToString
func EncodeBase64RawURL(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

// DecodeBase64Std base64.StdEncoding.DecodeString
func DecodeBase64Std(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// DecodeBase64URL base64.URLEncoding.DecodeString
func DecodeBase64URL(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// DecodeBase64RawURL base64.RawURLEncoding.DecodeString
func DecodeBase64RawURL(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
