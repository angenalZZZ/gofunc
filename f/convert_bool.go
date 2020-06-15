package f

import (
	"strconv"
	"strings"
)

// Bool convert string to bool, or return false.
func Bool(s string) bool {
	ok, _ := ToBool(s)
	return ok
}

// ToBool parse string to bool, or return ErrConvertFail.
// true ("1", "on", "yes", "true")
// false("0", "off", "no", "false")
func ToBool(s string) (bool, error) {
	lower := strings.ToLower(s)
	switch lower {
	case "1", "on", "yes", "true":
		return true, nil
	case "0", "off", "no", "false":
		return false, nil
	}
	return false, ErrConvertFail
}

// ToBoolean convert the input string to a boolean.
func ToBoolean(str string) (bool, error) {
	return strconv.ParseBool(str)
}
