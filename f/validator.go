package f

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Basic regular expressions for validating strings.
var (
	rxEmail          = regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
	rxISBN10         = regexp.MustCompile("^(?:[0-9]{9}X|[0-9]{10})$")
	rxISBN13         = regexp.MustCompile("^(?:[0-9]{13})$")
	rxUUID3          = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$")
	rxUUID4          = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	rxUUID5          = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$")
	rxUUID           = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
	rxAlpha          = regexp.MustCompile("^[a-zA-Z]+$")
	rxAlphaNum       = regexp.MustCompile("^[a-zA-Z0-9]+$")
	rxAlphaDash      = regexp.MustCompile(`^(?:[\w-]+)$`)
	rxNumber         = regexp.MustCompile("^[0-9]+$")
	rxInt            = regexp.MustCompile("^(?:[-+]?(?:0|[1-9][0-9]*))$")
	rxFloat          = regexp.MustCompile("^(?:[-+]?(?:[0-9]+))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$")
	rxCnMobile       = regexp.MustCompile(`^1\d{10}$`)
	rxHexColor       = regexp.MustCompile("^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$")
	rxRGBColor       = regexp.MustCompile("^rgb\\(\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*\\)$")
	rxASCII          = regexp.MustCompile("^[\x00-\x7F]+$")
	rxHexadecimal    = regexp.MustCompile("^[0-9a-fA-F]+$")
	rxPrintableASCII = regexp.MustCompile("^[\x20-\x7E]+$")
	rxMultiByte      = regexp.MustCompile("[^\x00-\x7F]")
	rxBase64         = regexp.MustCompile("^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$")
	rxDataURI        = regexp.MustCompile(`^data:.+/(.+);base64,(?:.+)`)
	rxLatitude       = regexp.MustCompile("^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$")
	rxLongitude      = regexp.MustCompile("^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$")
	rxDNSName        = regexp.MustCompile(`^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`)
	rxFullURL        = regexp.MustCompile(`^(?:ftp|tcp|udp|wss?|https?):\/\/[\w\.\/#=?&]+$`)
	rxURLSchema      = regexp.MustCompile(`((ftp|tcp|udp|wss?|https?):\/\/)`)
	rxWinPath        = regexp.MustCompile(`^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`)
	rxUnixPath       = regexp.MustCompile(`^(/[^/\x00]*)+/?$`)
	rxHasLowerCase   = regexp.MustCompile(".*[[:lower:]]")
	rxHasUpperCase   = regexp.MustCompile(".*[[:upper:]]")
)

var (
	emptyValue           = reflect.Value{}
	errConvertFail       = errors.New("convert value is failure")
	errBadComparisonType = errors.New("invalid type for operation")
)

// Contains asserts that the specified string, list(array, slice...) or map contains the
// specified substring or element.
//
//    Contains(t, "Hello World", "World")
//    Contains(t, ["Hello", "World"], "World")
//    Contains(t, {"Hello": "World"}, "Hello")
func Contains(s, sub interface{}) bool {
	ok, found := ContainsElement(s, sub)

	// ok == false: 's' could not be applied builtin len()
	// found == false: 's' does not contain 'sub'
	return ok && found
}

// NotContains check that the specified string, list(array, slice) or map does NOT contain the
// specified substring or element.
//
// Notice: list check value exist. map check key exist.
func NotContains(s, sub interface{}) bool {
	ok, found := ContainsElement(s, sub)

	// ok == false: could not be applied builtin len()
	// found == true: 's' contain 'sub'
	return ok && !found
}

// ContainsElement from package: github.com/stretchr/testify/assert/assertions.go
func ContainsElement(list, element interface{}) (ok, found bool) {
	listValue := reflect.ValueOf(list)
	listKind := reflect.TypeOf(list).Kind()
	defer func() {
		if e := recover(); e != nil {
			ok = false
			found = false
		}
	}()

	if listKind == reflect.String {
		elementValue := reflect.ValueOf(element)
		return true, strings.Contains(listValue.String(), elementValue.String())
	}

	if listKind == reflect.Map {
		mapKeys := listValue.MapKeys()
		for i := 0; i < len(mapKeys); i++ {
			if IsEqual(mapKeys[i].Interface(), element) {
				return true, true
			}
		}
		return true, false
	}

	for i := 0; i < listValue.Len(); i++ {
		if IsEqual(listValue.Index(i).Interface(), element) {
			return true, true
		}
	}
	return true, false
}

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

	intVal, err := ToInt(val, true)
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
	return s != "" && rxAlphaNum.MatchString(s)
}

// IsAlphaDash string.
func IsAlphaDash(s string) bool {
	return s != "" && rxAlphaDash.MatchString(s)
}

// IsNumber string. should >= 0
func IsNumber(v interface{}) bool {
	return rxNumber.MatchString(ToString(v))
}

// IsNumeric is string/int number. should >= 0
func IsNumeric(v interface{}) bool {
	return rxNumber.MatchString(ToString(v))
}

// IsStringNumber is string number. should >= 0
func IsStringNumber(s string) bool {
	return s != "" && rxNumber.MatchString(s)
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

// Eq evaluates the comparison a == b
// Support:
// 	bool, int(X), uint(X), string, float(X)
func Eq(arg1 reflect.Value, arg2 reflect.Value) (bool, error) {
	v1 := indirectInterface(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}

	v2 := indirectInterface(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}

	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
		}
		return truth, nil
	}

	switch k1 {
	case boolKind:
		truth = v1.Bool() == v2.Bool()
	case complexKind:
		truth = v1.Complex() == v2.Complex()
	case floatKind:
		truth = v1.Float() == v2.Float()
	case intKind:
		truth = v1.Int() == v2.Int()
	case stringKind:
		truth = v1.String() == v2.String()
	case uintKind:
		truth = v1.Uint() == v2.Uint()
	}

	return truth, nil
}

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}

	// like: slice, array, map ...
	return invalidKind, errBadComparisonType
}

// indirectInterface returns the concrete value in an interface value,
// or else the zero reflect.Value.
// That is, if v represents the interface value x, the result is the same as reflect.ValueOf(x):
// the fact that x was an interface value is forgotten.
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}

	if v.IsNil() {
		return emptyValue
	}

	return v.Elem()
}

// containsKind checks if a specified kind in the slice of kinds.
func containsKind(kinds []reflect.Kind, kind reflect.Kind) bool {
	for i := 0; i < len(kinds); i++ {
		if kind == kinds[i] {
			return true
		}
	}

	return false
}

/*************************************************************
 * global: length validators
 *************************************************************/

// ValueLen get value length
func ValueLen(v reflect.Value) int {
	k := v.Kind()

	// (u)int use width.
	switch k {
	case reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.String:
		return v.Len()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return len(fmt.Sprint(v.Uint()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return len(fmt.Sprint(v.Int()))
	case reflect.Float32, reflect.Float64:
		return len(fmt.Sprint(v.Interface()))
	}

	// cannot get length
	return -1
}

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

/*************************************************************
 * Reflection:
 * From package(go 1.13) "reflect" -> reflect/value.go
 *************************************************************/

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
// NOTICE: this's an built-in method in reflect/value.go since go 1.13
func IsZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return math.Float64bits(v.Float()) == 0
	case reflect.Complex64, reflect.Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			// if !v.Index(i).IsZero() {
			if !IsZero(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			// if !v.Index(i).IsZero() {
			if !IsZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// This should never happens, but will act as a safeguard for
		// later, as a default value doesn't makes sense here.
		panic(&reflect.ValueError{Method: "cannot check reflect.Value.IsZero", Kind: v.Kind()})
	}
}

// IsEmpty Is the variable empty.
func IsEmpty(val interface{}) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Chan:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	}

	return reflect.DeepEqual(val, reflect.Zero(v.Type()).Interface())
}

// IsNil checks if a specified object is nil or not, without Failing.
func IsNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	isNilKind := containsKind(
		[]reflect.Kind{
			reflect.Chan, reflect.Func,
			reflect.Interface, reflect.Map,
			reflect.Ptr, reflect.Slice},
		kind)

	if isNilKind && value.IsNil() {
		return true
	}

	return false
}
