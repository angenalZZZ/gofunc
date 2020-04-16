package hex

import "strconv"

// ParseInt decodes a hex string as a quantity.
func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 16, 64)
}

// ParseUint decodes a hex string as a quantity.
func ParseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 16, 64)
}

// FormatInt encodes i as a hex string.
func FormatInt(i int64) string {
	return strconv.FormatInt(i, 16)
}

// FormatUint encodes i as a hex string.
func FormatUint(i uint64) string {
	return strconv.FormatUint(i, 16)
}
