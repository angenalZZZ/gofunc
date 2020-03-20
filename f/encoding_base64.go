package f

import "encoding/base64"

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
