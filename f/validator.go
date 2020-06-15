package f

import (
	"bytes"
	"encoding/json"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// IsUint check, allow: intX, uintX, string
func IsUint(val interface{}) bool {
	switch typVal := val.(type) {
	case int:
		return typVal >= 0
	case int8:
		return typVal >= 0
	case int16:
		return typVal >= 0
	case int32:
		return typVal >= 0
	case int64:
		return typVal >= 0
	case uint, uint8, uint16, uint32, uint64:
		return true
	case string:
		_, err := strconv.ParseUint(typVal, 10, 32)
		return err == nil
	}
	return false
}

// IsBool check. allow: bool, string.
func IsBool(val interface{}) bool {
	if _, ok := val.(bool); ok {
		return true
	}

	if typVal, ok := val.(string); ok {
		_, err := ToBool(typVal)
		return err == nil
	}
	return false
}

// IsFloat check. allow: floatX, string
func IsFloat(val interface{}) bool {
	if val == nil {
		return false
	}

	switch rv := val.(type) {
	case float32, float64:
		return true
	case string:
		return rv != "" && rxFloat.MatchString(rv)
	}
	return false
}

// IsArray check
func IsArray(val interface{}) (ok bool) {
	if val == nil {
		return false
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv.Kind() == reflect.Array
}

// IsSlice check
func IsSlice(val interface{}) (ok bool) {
	if val == nil {
		return false
	}

	rv := reflect.ValueOf(val)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Slice
}

// IsIntSlice is int slice check
func IsIntSlice(val interface{}) bool {
	if val == nil {
		return false
	}

	switch val.(type) {
	case []int, []int8, []int16, []int32, []int64, []uint, []uint8, []uint16, []uint32, []uint64:
		return true
	}
	return false
}

// IsStrings is string slice check
func IsStrings(val interface{}) (ok bool) {
	if val == nil {
		return false
	}

	_, ok = val.([]string)
	return
}

// IsMap check
func IsMap(val interface{}) (ok bool) {
	if val == nil {
		return false
	}

	var rv reflect.Value
	if rv, ok = val.(reflect.Value); !ok {
		rv = reflect.ValueOf(val)
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv.Kind() == reflect.Map
}

// IsInt check, and support length check
func IsInt(val interface{}, minAndMax ...int64) (ok bool) {
	if val == nil {
		return false
	}

	intVal, err := ToInt(val)
	if err != nil {
		return false
	}

	argLn := len(minAndMax)
	if argLn == 0 { // only check type
		return true
	}

	// value check
	minVal := minAndMax[0]
	if argLn == 1 { // only min length check.
		return intVal >= minVal
	}

	maxVal := minAndMax[1]

	// min and max length check
	return intVal >= minVal && intVal <= maxVal
}

// IsString check and support length check.
// Usage:
// 	ok := IsString(val)
// 	ok := IsString(val, 5) // with min len check
// 	ok := IsString(val, 5, 12) // with min and max len check
func IsString(val interface{}, minAndMaxLen ...int) (ok bool) {
	if val == nil {
		return false
	}

	argLn := len(minAndMaxLen)
	str, isStr := val.(string)

	// only check type
	if argLn == 0 {
		return isStr
	}

	if !isStr {
		return false
	}

	// length check
	strLen := len(str)
	minLen := minAndMaxLen[0]

	// only min length check.
	if argLn == 1 {
		return strLen >= minLen
	}

	// min and max length check
	maxLen := minAndMaxLen[1]
	return strLen >= minLen && strLen <= maxLen
}

/*************************************************************
 * global: string validators
 *************************************************************/

// HasWhitespace check. eg "10"
func HasWhitespace(s string) bool {
	return s != "" && strings.ContainsRune(s, ' ')
}

// IsIntString check. eg "10"
func IsIntString(s string) bool {
	return s != "" && rxInt.MatchString(s)
}

// IsASCII string.
func IsASCII(s string) bool {
	return s != "" && rxASCII.MatchString(s)
}

// IsPrintableASCII string.
func IsPrintableASCII(s string) bool {
	return s != "" && rxPrintableASCII.MatchString(s)
}

// IsBase64 string.
func IsBase64(s string) bool {
	return s != "" && rxBase64.MatchString(s)
}

// IsLatitude string.
func IsLatitude(s string) bool {
	return s != "" && rxLatitude.MatchString(s)
}

// IsLongitude string.
func IsLongitude(s string) bool {
	return s != "" && rxLongitude.MatchString(s)
}

// IsDNSName string.
func IsDNSName(s string) bool {
	return s != "" && rxDNSName.MatchString(s)
}

// HasURLSchema string.
func HasURLSchema(s string) bool {
	return s != "" && rxURLSchema.MatchString(s)
}

// IsFullURL string.
func IsFullURL(s string) bool {
	return s != "" && rxFullURL.MatchString(s)
}

// IsURL string.
func IsURL(s string) bool {
	if s == "" {
		return false
	}

	_, err := url.Parse(s)
	return err == nil
}

// IsDataURI string.
// data:[<mime type>] ( [;charset=<charset>] ) [;base64],码内容
// eg. "data:image/gif;base64,R0lGODlhA..."
func IsDataURI(s string) bool {
	return s != "" && rxDataURI.MatchString(s)
}

// IsMultiByte string.
func IsMultiByte(s string) bool {
	return s != "" && rxMultiByte.MatchString(s)
}

// IsISBN10 string.
func IsISBN10(s string) bool {
	return s != "" && rxISBN10.MatchString(s)
}

// IsISBN13 string.
func IsISBN13(s string) bool {
	return s != "" && rxISBN13.MatchString(s)
}

// IsHexadecimal string.
func IsHexadecimal(s string) bool {
	return s != "" && rxHexadecimal.MatchString(s)
}

// IsCnMobile string.
func IsCnMobile(s string) bool {
	return s != "" && rxCnMobile.MatchString(s)
}

// IsHexColor string.
func IsHexColor(s string) bool {
	return s != "" && rxHexColor.MatchString(s)
}

// IsRGBColor string.
func IsRGBColor(s string) bool {
	return s != "" && rxRGBColor.MatchString(s)
}

// IsAlpha string.
func IsAlpha(s string) bool {
	return s != "" && rxAlpha.MatchString(s)
}

// IsAlphaNum string.
func IsAlphaNum(s string) bool {
	return s != "" && rxAlphaNumeric.MatchString(s)
}

// IsAlphaDash string.
func IsAlphaDash(s string) bool {
	return s != "" && rxAlphaDash.MatchString(s)
}

// IsNumber string. should >= 0
func IsNumber(v interface{}) bool {
	return rxNumeric.MatchString(ToString(v))
}

// IsNumeric is string/int number. should >= 0
func IsNumeric(v interface{}) bool {
	return rxNumeric.MatchString(ToString(v))
}

// IsStringNumber is string number. should >= 0
func IsStringNumber(s string) bool {
	return s != "" && rxNumeric.MatchString(s)
}

// IsEmail check
func IsEmail(s string) bool {
	return s != "" && rxEmail.MatchString(s)
}

// IsUUID string
func IsUUID(s string) bool {
	return s != "" && rxUUID.MatchString(s)
}

// IsUUID3 string
func IsUUID3(s string) bool {
	return s != "" && rxUUID3.MatchString(s)
}

// IsUUID4 string
func IsUUID4(s string) bool {
	return s != "" && rxUUID4.MatchString(s)
}

// IsUUID5 string
func IsUUID5(s string) bool {
	return s != "" && rxUUID5.MatchString(s)
}

// IsIP is the validation function for validating if the field's value is a valid v4 or v6 IP address.
func IsIP(s string) bool {
	// ip := net.ParseIP(s)
	return s != "" && net.ParseIP(s) != nil
}

// IsIPv4 is the validation function for validating if a value is a valid v4 IP address.
func IsIPv4(s string) bool {
	if s == "" {
		return false
	}

	ip := net.ParseIP(s)
	return ip != nil && ip.To4() != nil
}

// IsIPv6 is the validation function for validating if the field's value is a valid v6 IP address.
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && ip.To4() == nil
}

// IsMAC is the validation function for validating if the field's value is a valid MAC address.
func IsMAC(s string) bool {
	if s == "" {
		return false
	}
	_, err := net.ParseMAC(s)
	return err == nil
}

// IsCIDRv4 is the validation function for validating if the field's value is a valid v4 CIDR address.
func IsCIDRv4(s string) bool {
	if s == "" {
		return false
	}
	ip, _, err := net.ParseCIDR(s)
	return err == nil && ip.To4() != nil
}

// IsCIDRv6 is the validation function for validating if the field's value is a valid v6 CIDR address.
func IsCIDRv6(s string) bool {
	if s == "" {
		return false
	}

	ip, _, err := net.ParseCIDR(s)
	return err == nil && ip.To4() == nil
}

// IsCIDR is the validation function for validating if the field's value is a valid v4 or v6 CIDR address.
func IsCIDR(s string) bool {
	if s == "" {
		return false
	}

	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// IsJSON check if the string is valid JSON (note: uses json.Unmarshal).
func IsJSON(s string) bool {
	if s == "" {
		return false
	}

	var js json.RawMessage
	return DecodeJson([]byte(s), &js) == nil
}

// HasLowerCase check string has lower case
func HasLowerCase(s string) bool {
	if s == "" {
		return false
	}

	return rxHasLowerCase.MatchString(s)
}

// HasUpperCase check string has upper case
func HasUpperCase(s string) bool {
	if s == "" {
		return false
	}

	return rxHasUpperCase.MatchString(s)
}

// StartsWith check string is starts with sub-string
func StartsWith(s, sub string) bool {
	if s == "" {
		return false
	}

	return strings.HasPrefix(s, sub)
}

// EndsWith check string is ends with sub-string
func EndsWith(s, sub string) bool {
	if s == "" {
		return false
	}

	return strings.HasSuffix(s, sub)
}

// StringContains check string is contains sub-string
func StringContains(s, sub string) bool {
	if s == "" {
		return false
	}

	return strings.Contains(s, sub)
}

// Regexp match value string
func Regexp(str string, pattern string) bool {
	ok, _ := regexp.MatchString(pattern, str)
	return ok
}

/*************************************************************
 * global: filesystem validators
 *************************************************************/

// PathExists reports whether the named file or directory exists.
func PathExists(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// IsFilePath string
func IsFilePath(path string) bool {
	if path == "" {
		return false
	}

	if fi, err := os.Stat(path); err == nil {
		return !fi.IsDir()
	}
	return false
}

func IsDirPath(path string) bool {
	if path == "" {
		return false
	}

	if fi, err := os.Stat(path); err == nil {
		return fi.IsDir()
	}
	return false
}

// IsWinPath string
func IsWinPath(s string) bool {
	return s != "" && rxWinPath.MatchString(s)
}

// IsUnixPath string
func IsUnixPath(s string) bool {
	return s != "" && rxUnixPath.MatchString(s)
}

/*************************************************************
 * global: compare validators
 *************************************************************/

// Implements asserts that an object is implemented by the specified interface.
//
//    Implements((*MyInterface)(nil), new(MyObject))
func Implements(interfaceObject interface{}, object interface{}) bool {
	if object == nil {
		return false
	}

	interfaceType := reflect.TypeOf(interfaceObject).Elem()
	return reflect.TypeOf(object).Implements(interfaceType)
}

// AreEqual determines if two objects are considered equal.
//
// This function does no assertion of any kind.
func AreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}

// AreEqualValues gets whether two objects are equal, or if their
// values are equal.
func AreEqualValues(expected, actual interface{}) bool {
	if AreEqual(expected, actual) {
		return true
	}

	actualType := reflect.TypeOf(actual)
	if actualType == nil {
		return false
	}
	expectedValue := reflect.ValueOf(expected)
	if expectedValue.IsValid() && expectedValue.Type().ConvertibleTo(actualType) {
		// Attempt comparison after type conversion
		return reflect.DeepEqual(expectedValue.Convert(actualType).Interface(), actual)
	}

	return false
}

// IsEqual check two value is equals.
// Support:
// 	bool, int(X), uint(X), string, float(X) AND slice, array, map
func IsEqual(val, wantVal interface{}) bool {
	// check is nil
	if val == nil || wantVal == nil {
		return val == wantVal
	}

	sv := reflect.ValueOf(val)
	wv := reflect.ValueOf(wantVal)

	// don't compare func, struct
	if sv.Kind() == reflect.Func || sv.Kind() == reflect.Struct {
		return false
	}
	if wv.Kind() == reflect.Func || wv.Kind() == reflect.Struct {
		return false
	}

	// compare basic type: bool, int(X), uint(X), string, float(X)
	equal, err := Eq(sv, wv)

	// is not an basic type, eg: slice, array, map ...
	if err != nil {
		expBt, ok := val.([]byte)
		if !ok {
			return reflect.DeepEqual(val, wantVal)
		}

		actBt, ok := wantVal.([]byte)
		if !ok {
			return false
		}
		if expBt == nil || actBt == nil {
			return expBt == nil && actBt == nil
		}

		return bytes.Equal(expBt, actBt)
	}

	return equal
}

// NotEqual check
func NotEqual(val, wantVal interface{}) bool {
	return !IsEqual(val, wantVal)
}

/*************************************************************
 * global: length validators
 *************************************************************/

// CalcLength for input value
func CalcLength(val interface{}) int {
	if val == nil {
		return -1
	}

	// string length
	if str, ok := val.(string); ok {
		return len(str)
	}
	return ValueLen(reflect.ValueOf(val))
}

// Length equal check for string, array, slice, map
func Length(val interface{}, wantLen int) bool {
	ln := CalcLength(val)
	if ln == -1 {
		return false
	}

	return ln == wantLen
}

// MinLength check for string, array, slice, map
func MinLength(val interface{}, minLen int) bool {
	ln := CalcLength(val)
	if ln == -1 {
		return false
	}

	return ln >= minLen
}

// MaxLength check for string, array, slice, map
func MaxLength(val interface{}, maxLen int) bool {
	ln := CalcLength(val)
	if ln == -1 {
		return false
	}

	return ln <= maxLen
}

// ByteLength check string's length
func ByteLength(str string, minLen int, maxLen ...int) bool {
	strLen := len(str)

	// only min length check.
	if len(maxLen) == 0 {
		return strLen >= minLen
	}

	// min and max length check
	return strLen >= minLen && strLen <= maxLen[0]
}

// RuneLength check string's length (including multi byte strings)
func RuneLength(val interface{}, minLen int, maxLen ...int) bool {
	str, isString := val.(string)
	if !isString {
		return false
	}

	// strLen := len([]rune(str))
	strLen := utf8.RuneCountInString(str)

	// only min length check.
	if len(maxLen) == 0 {
		return strLen >= minLen
	}

	// min and max length check
	return strLen >= minLen && strLen <= maxLen[0]
}

// StringLength check string's length (including multi byte strings)
func StringLength(val interface{}, minLen int, maxLen ...int) bool {
	return RuneLength(val, minLen, maxLen...)
}
