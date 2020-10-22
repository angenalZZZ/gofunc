package f

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"html"
	"io/ioutil"
	"math"
	"net"
	"net/url"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Contains asserts that the specified string, list(array, slice...) or map contains the
// specified substring or element.
//
//    Contains("Hello World", "World")
//    Contains(["Hello", "World"], "World")
//    Contains({"Hello": "World"}, "Hello")
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

// Matches checks if string matches the pattern (pattern is regular expression)
// In case of error return false
func Matches(str, pattern string) bool {
	match, _ := regexp.MatchString(pattern, str)
	return match
}

// LeftTrim trims characters from the left side of the input.
// If second argument is empty, it will remove leading spaces.
func LeftTrim(str, chars string) string {
	if chars == "" {
		return strings.TrimLeftFunc(str, unicode.IsSpace)
	}
	r, _ := regexp.Compile("^[" + chars + "]+")
	return r.ReplaceAllString(str, "")
}

// RightTrim trims characters from the right side of the input.
// If second argument is empty, it will remove trailing spaces.
func RightTrim(str, chars string) string {
	if chars == "" {
		return strings.TrimRightFunc(str, unicode.IsSpace)
	}
	r, _ := regexp.Compile("[" + chars + "]+$")
	return r.ReplaceAllString(str, "")
}

// Trim trims characters from both sides of the input.
// If second argument is empty, it will remove spaces.
func Trim(str, chars string) string {
	return LeftTrim(RightTrim(str, chars), chars)
}

// WhiteList removes characters that do not appear in the whitelist.
func WhiteList(str, chars string) string {
	pattern := "[^" + chars + "]+"
	r, _ := regexp.Compile(pattern)
	return r.ReplaceAllString(str, "")
}

// BlackList removes characters that appear in the blacklist.
func BlackList(str, chars string) string {
	pattern := "[" + chars + "]+"
	r, _ := regexp.Compile(pattern)
	return r.ReplaceAllString(str, "")
}

// StripLow removes characters with a numerical value < 32 and 127, mostly control characters.
// If keep_new_lines is true, newline characters are preserved (\n and \r, hex 0xA and 0xD).
func StripLow(str string, keepNewLines bool) string {
	chars := ""
	if keepNewLines {
		chars = "\x00-\x09\x0B\x0C\x0E-\x1F\x7F"
	} else {
		chars = "\x00-\x1F\x7F"
	}
	return BlackList(str, chars)
}

// ReplacePattern replaces regular expression pattern in string
func ReplacePattern(str, pattern, replace string) string {
	r, _ := regexp.Compile(pattern)
	return r.ReplaceAllString(str, replace)
}

// Escape replaces <, >, & and " with HTML entities.
var Escape = html.EscapeString

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

// UnderscoreToCamelCase converts from underscore separated form to camel case form.
// Ex.: my_func => MyFunc
func UnderscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

// CamelCaseToUnderscore converts from camel case form to underscore separated form.
// Ex.: MyFunc => my_func
func CamelCaseToUnderscore(str string) string {
	var output []rune
	var segment []rune
	for _, r := range str {

		// not treat number as separate segment
		if !unicode.IsLower(r) && string(r) != "_" && !unicode.IsNumber(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

// Reverse returns reversed string
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// GetLines splits string by "\n" and return array of lines
func GetLines(s string) []string {
	return strings.Split(s, "\n")
}

// GetLine returns specified line of multiline string
func GetLine(s string, index int) (string, error) {
	lines := GetLines(s)
	if index < 0 || index >= len(lines) {
		return "", errors.New("line index out of bounds")
	}
	return lines[index], nil
}

// RemoveTags removes all tags from HTML string
func RemoveTags(s string) string {
	return ReplacePattern(s, "<[^>]*>", "")
}

// SafeFileName returns safe string that can be used in file names
func SafeFileName(str string) string {
	name := strings.ToLower(str)
	name = path.Clean(path.Base(name))
	name = strings.Trim(name, " ")
	separators, err := regexp.Compile(`[ &_=+:]`)
	if err == nil {
		name = separators.ReplaceAllString(name, "-")
	}
	legal, err := regexp.Compile(`[^[:alnum:]-.]`)
	if err == nil {
		name = legal.ReplaceAllString(name, "")
	}
	for strings.Contains(name, "--") {
		name = strings.Replace(name, "--", "-", -1)
	}
	return name
}

// NormalizeEmail canonicalize an email address.
// The local part of the email address is lowercased for all domains; the hostname is always lowercased and
// the local part of the email address is always lowercased for hosts that are known to be case-insensitive (currently only GMail).
// Normalization follows special rules for known providers: currently, GMail addresses have dots removed in the local part and
// are stripped of tags (e.g. some.one+tag@gmail.com becomes someone@gmail.com) and all @googlemail.com addresses are
// normalized to @gmail.com.
func NormalizeEmail(str string) (string, error) {
	if !IsEmail(str) {
		return "", fmt.Errorf("%s is not an email", str)
	}
	parts := strings.Split(str, "@")
	parts[0] = strings.ToLower(parts[0])
	parts[1] = strings.ToLower(parts[1])
	if parts[1] == "gmail.com" || parts[1] == "googlemail.com" {
		parts[1] = "gmail.com"
		parts[0] = strings.Split(ReplacePattern(parts[0], `\.`, ""), "+")[0]
	}
	return strings.Join(parts, "@"), nil
}

// Truncate a string to the closest length without breaking words.
func Truncate(str string, length int, ending string) string {
	var aftstr, befstr string
	if len(str) > length {
		words := strings.Fields(str)
		before, present := 0, 0
		for i := range words {
			befstr = aftstr
			before = present
			aftstr = aftstr + words[i] + " "
			present = len(aftstr)
			if present > length && i != 0 {
				if (length - before) < (present - length) {
					return Trim(befstr, " /\\.,\"'#!?&@+-") + ending
				}
				return Trim(aftstr, " /\\.,\"'#!?&@+-") + ending
			}
		}
	}

	return str
}

// PadLeft pads left side of a string if size of string is less then indicated pad length
func PadLeft(str string, padStr string, padLen int) string {
	return buildPadStr(str, padStr, padLen, true, false)
}

// PadRight pads right side of a string if size of string is less then indicated pad length
func PadRight(str string, padStr string, padLen int) string {
	return buildPadStr(str, padStr, padLen, false, true)
}

// PadBoth pads both sides of a string if size of string is less then indicated pad length
func PadBoth(str string, padStr string, padLen int) string {
	return buildPadStr(str, padStr, padLen, true, true)
}

// PadString either left, right or both sides.
// Note that padding string can be unicode and more then one character
func buildPadStr(str string, padStr string, padLen int, padLeft bool, padRight bool) string {

	// When padded length is less then the current string size
	if padLen < utf8.RuneCountInString(str) {
		return str
	}

	padLen -= utf8.RuneCountInString(str)

	targetLen := padLen

	targetLenLeft := targetLen
	targetLenRight := targetLen
	if padLeft && padRight {
		targetLenLeft = padLen / 2
		targetLenRight = padLen - targetLenLeft
	}

	strToRepeatLen := utf8.RuneCountInString(padStr)

	repeatTimes := int(math.Ceil(float64(targetLen) / float64(strToRepeatLen)))
	repeatedString := strings.Repeat(padStr, repeatTimes)

	leftSide := ""
	if padLeft {
		leftSide = repeatedString[0:targetLenLeft]
	}

	rightSide := ""
	if padRight {
		rightSide = repeatedString[0:targetLenRight]
	}

	return leftSide + str + rightSide
}

// TruncatingErrorf removes extra args from fmt.Errorf if not formatted in the str object
func TruncatingErrorf(str string, args ...interface{}) error {
	n := strings.Count(str, "%s")
	return fmt.Errorf(str, args[:n]...)
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

// IsFloat check if the string is a float.
func IsFloat(str string) bool {
	return str != "" && rxFloat.MatchString(str)
}

// IsArray check
func IsArray(val interface{}) bool {
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
func IsSlice(val interface{}) bool {
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
func IsStrings(val interface{}) bool {
	if val == nil {
		return false
	}

	if _, ok := val.([]string); ok {
		return true
	}
	//if _, ok := val.(g.Strings); ok {
	//	return true
	//}
	return false
}

// IsMap check
func IsMap(val interface{}) (ok bool) {
	if val == nil {
		return false
	}

	//if _, ok = val.(g.Map); ok {
	//	return true
	//}
	var rv reflect.Value
	if rv, ok = val.(reflect.Value); !ok {
		rv = reflect.ValueOf(val)
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	return rv.Kind() == reflect.Map
}

// IsInt check if the string is an integer. Empty string is not valid.
func IsInt(str string) bool {
	if IsNull(str) {
		return false
	}
	return rxInt.MatchString(str)
}

// IsInt check, and support length check
func IsInt2(val interface{}, minAndMax ...int64) (ok bool) {
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

// HasWhitespace checks if the string contains any whitespace
func HasWhitespace(str string) bool {
	return len(str) > 0 && rxHasWhitespace.MatchString(str)
}

// HasWhitespace check. eg "10"
func HasWhitespace2(s string) bool {
	return s != "" && strings.ContainsRune(s, ' ')
}

// IsIntString check. eg "10"
func IsIntString(s string) bool {
	return s != "" && rxInt.MatchString(s)
}

// IsASCII check if the string contains ASCII chars only. Empty string is valid.
func IsASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxASCII.MatchString(str)
}

// IsPrintableASCII check if the string contains printable ASCII chars only. Empty string is valid.
func IsPrintableASCII(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxPrintableASCII.MatchString(str)
}

// IsFullWidth check if the string contains any full-width chars. Empty string is valid.
func IsFullWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxFullWidth.MatchString(str)
}

// IsHalfWidth check if the string contains any half-width chars. Empty string is valid.
func IsHalfWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str)
}

// IsVariableWidth check if the string contains a mixture of full and half-width chars. Empty string is valid.
func IsVariableWidth(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHalfWidth.MatchString(str) && rxFullWidth.MatchString(str)
}

// IsBase64 check if a string is base64 encoded.
func IsBase64(s string) bool {
	return s != "" && rxBase64.MatchString(s)
}

// IsLatitude check if a string is valid latitude.
func IsLatitude(s string) bool {
	return s != "" && rxLatitude.MatchString(s)
}

// IsLongitude check if a string is valid longitude.
func IsLongitude(s string) bool {
	return s != "" && rxLongitude.MatchString(s)
}

// IsIMEI check if a string is valid IMEI
func IsIMEI(str string) bool {
	return rxIMEI.MatchString(str)
}

// IsRsaPublicKey check if a string is valid public key with provided length
func IsRsaPublicKey(str string, keylen int) bool {
	bb := bytes.NewBufferString(str)
	pemBytes, err := ioutil.ReadAll(bb)
	if err != nil {
		return false
	}
	block, _ := pem.Decode(pemBytes)
	if block != nil && block.Type != "PUBLIC KEY" {
		return false
	}
	var der []byte

	if block != nil {
		der = block.Bytes
	} else {
		der, err = base64.StdEncoding.DecodeString(str)
		if err != nil {
			return false
		}
	}

	key, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return false
	}
	pubkey, ok := key.(*rsa.PublicKey)
	if !ok {
		return false
	}
	bitlen := len(pubkey.N.Bytes()) * 8
	return bitlen == int(keylen)
}

func toJSONName(tag string) string {
	if tag == "" {
		return ""
	}

	// JSON name always comes first. If there's no options then split[0] is
	// JSON name, if JSON name is not set, then split[0] is an empty string.
	split := strings.SplitN(tag, ",", 2)

	name := split[0]

	// However it is possible that the field is skipped when
	// (de-)serializing from/to JSON, in which case assume that there is no
	// tag name to use
	if name == "-" {
		return ""
	}
	return name
}

func PrependPathToErrors(err error, path string) error {
	switch err2 := err.(type) {
	case Error:
		err2.Path = append([]string{path}, err2.Path...)
		return err2
	case Errors:
		errors := err2.Errors()
		for i, err3 := range errors {
			errors[i] = PrependPathToErrors(err3, path)
		}
		return err2
	}
	return err
}

// IsHash checks if a string is a hash of type algorithm.
// Algorithm is one of ['md4', 'md5', 'sha1', 'sha256', 'sha384', 'sha512', 'ripemd128', 'ripemd160', 'tiger128', 'tiger160', 'tiger192', 'crc32', 'crc32b']
func IsHash(str string, algorithm string) bool {
	len := "0"
	algo := strings.ToLower(algorithm)

	if algo == "crc32" || algo == "crc32b" {
		len = "8"
	} else if algo == "md5" || algo == "md4" || algo == "ripemd128" || algo == "tiger128" {
		len = "32"
	} else if algo == "sha1" || algo == "ripemd160" || algo == "tiger160" {
		len = "40"
	} else if algo == "tiger192" {
		len = "48"
	} else if algo == "sha256" {
		len = "64"
	} else if algo == "sha384" {
		len = "96"
	} else if algo == "sha512" {
		len = "128"
	} else {
		return false
	}

	return Matches(str, "^[a-f0-9]{"+len+"}$")
}

// IsSHA512 checks is a string is a SHA512 hash. Alias for `IsHash(str, "sha512")`
func IsSHA512(str string) bool {
	return IsHash(str, "sha512")
}

// IsSHA384 checks is a string is a SHA384 hash. Alias for `IsHash(str, "sha384")`
func IsSHA384(str string) bool {
	return IsHash(str, "sha384")
}

// IsSHA256 checks is a string is a SHA256 hash. Alias for `IsHash(str, "sha256")`
func IsSHA256(str string) bool {
	return IsHash(str, "sha256")
}

// IsTiger192 checks is a string is a Tiger192 hash. Alias for `IsHash(str, "tiger192")`
func IsTiger192(str string) bool {
	return IsHash(str, "tiger192")
}

// IsTiger160 checks is a string is a Tiger160 hash. Alias for `IsHash(str, "tiger160")`
func IsTiger160(str string) bool {
	return IsHash(str, "tiger160")
}

// IsRipeMD160 checks is a string is a RipeMD160 hash. Alias for `IsHash(str, "ripemd160")`
func IsRipeMD160(str string) bool {
	return IsHash(str, "ripemd160")
}

// IsSHA1 checks is a string is a SHA-1 hash. Alias for `IsHash(str, "sha1")`
func IsSHA1(str string) bool {
	return IsHash(str, "sha1")
}

// IsTiger128 checks is a string is a Tiger128 hash. Alias for `IsHash(str, "tiger128")`
func IsTiger128(str string) bool {
	return IsHash(str, "tiger128")
}

// IsRipeMD128 checks is a string is a RipeMD128 hash. Alias for `IsHash(str, "ripemd128")`
func IsRipeMD128(str string) bool {
	return IsHash(str, "ripemd128")
}

// IsCRC32 checks is a string is a CRC32 hash. Alias for `IsHash(str, "crc32")`
func IsCRC32(str string) bool {
	return IsHash(str, "crc32")
}

// IsCRC32b checks is a string is a CRC32b hash. Alias for `IsHash(str, "crc32b")`
func IsCRC32b(str string) bool {
	return IsHash(str, "crc32b")
}

// IsMD5 checks is a string is a MD5 hash. Alias for `IsHash(str, "md5")`
func IsMD5(str string) bool {
	return IsHash(str, "md5")
}

// IsMD4 checks is a string is a MD4 hash. Alias for `IsHash(str, "md4")`
func IsMD4(str string) bool {
	return IsHash(str, "md4")
}

// IsDialString validates the given string for usage with the various Dial() functions
func IsDialString(str string) bool {
	if h, p, err := net.SplitHostPort(str); err == nil && h != "" && p != "" && (IsDNSName(h) || IsIP(h)) && IsPort(p) {
		return true
	}

	return false
}

// IsDNSName will validate the given string as a DNS name
func IsDNSName(str string) bool {
	if str == "" || len(strings.Replace(str, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return !IsIP(str) && rxDNSName.MatchString(str)
}

// HasURLSchema string.
func HasURLSchema(s string) bool {
	return s != "" && rxURLSchema.MatchString(s)
}

// IsRequestURL check if the string rawurl, assuming
// it was received in an HTTP request, is a valid
// URL confirm to RFC 3986
func IsRequestURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false //Couldn't even parse the rawurl
	}
	if len(u.Scheme) == 0 {
		return false //No Scheme found
	}
	return true
}

// IsRequestURI check if the string rawurl, assuming
// it was received in an HTTP request, is an
// absolute URI or an absolute path.
func IsRequestURI(s string) bool {
	_, err := url.ParseRequestURI(s)
	return err == nil
}

// IsFullURL string.
func IsFullURL(s string) bool {
	return s != "" && rxFullURL.MatchString(s)
}

// IsURL check if the string is an URL.
func IsURL(str string) bool {
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(str)
}

// IsDataURI checks if a string is base64 encoded data URI such as an image
func IsDataURI(str string) bool {
	dataURI := strings.Split(str, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return IsBase64(dataURI[1])
}

// IsMagnetURI checks if a string is valid magnet URI
func IsMagnetURI(str string) bool {
	return rxMagnetURI.MatchString(str)
}

// IsISO3166Alpha2 checks if a string is valid two-letter country code
func IsISO3166Alpha2(str string) bool {
	for _, entry := range ISO3166List {
		if str == entry.Alpha2Code {
			return true
		}
	}
	return false
}

// IsISO3166Alpha3 checks if a string is valid three-letter country code
func IsISO3166Alpha3(str string) bool {
	for _, entry := range ISO3166List {
		if str == entry.Alpha3Code {
			return true
		}
	}
	return false
}

// IsISO693Alpha2 checks if a string is valid two-letter language code
func IsISO693Alpha2(str string) bool {
	for _, entry := range ISO693List {
		if str == entry.Alpha2Code {
			return true
		}
	}
	return false
}

// IsISO693Alpha3b checks if a string is valid three-letter language code
func IsISO693Alpha3b(str string) bool {
	for _, entry := range ISO693List {
		if str == entry.Alpha3bCode {
			return true
		}
	}
	return false
}

// IsMultiByte check if the string contains one or more multibyte chars. Empty string is valid.
func IsMultiByte(s string) bool {
	if IsNull(s) {
		return true
	}
	return s != "" && rxMultiByte.MatchString(s)
}

// IsISBN10 string.
func IsISBN10(s string) bool {
	return s != "" && IsISBN(s, 10)
}

// IsISBN13 string.
func IsISBN13(s string) bool {
	return s != "" && IsISBN(s, 13)
}

// IsISBN check if the string is an ISBN (version 10 or 13).
// If version value is not equal to 10 or 13, it will be check both variants.
func IsISBN(str string, version int) bool {
	sanitized := whiteSpacesAndMinus.ReplaceAllString(str, "")
	var checksum int32
	var i int32
	if version == 10 {
		if !rxISBN10.MatchString(sanitized) {
			return false
		}
		for i = 0; i < 9; i++ {
			checksum += (i + 1) * int32(sanitized[i]-'0')
		}
		if sanitized[9] == 'X' {
			checksum += 10 * 10
		} else {
			checksum += 10 * int32(sanitized[9]-'0')
		}
		if checksum%11 == 0 {
			return true
		}
		return false
	} else if version == 13 {
		if !rxISBN13.MatchString(sanitized) {
			return false
		}
		factor := []int32{1, 3}
		for i = 0; i < 12; i++ {
			checksum += factor[i%2] * int32(sanitized[i]-'0')
		}
		return (int32(sanitized[12]-'0'))-((10-(checksum%10))%10) == 0
	}
	return IsISBN(str, 10) || IsISBN(str, 13)
}

// IsHexadecimal check if the string is a hexadecimal number.
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

// IsLowerCase check if the string is lowercase. Empty string is valid.
func IsLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToLower(str)
}

// IsUpperCase check if the string is uppercase. Empty string is valid.
func IsUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return str == strings.ToUpper(str)
}

//IsUTFLetter check if the string contains only unicode letter characters.
//Similar to IsAlpha but for all languages. Empty string is valid.
func IsUTFLetter(str string) bool {
	if IsNull(str) {
		return true
	}

	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true

}

// IsAlphanumeric check if the string contains only letters and numbers. Empty string is valid.
func IsAlphanumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlphaNumeric.MatchString(str)
}

// IsUTFLetterNumeric check if the string contains only unicode letters and numbers. Empty string is valid.
func IsUTFLetterNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	for _, c := range str {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) { //letters && numbers are ok
			return false
		}
	}
	return true
}

// IsAlpha check if the string contains only letters (a-zA-Z). Empty string is valid.
func IsAlpha(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxAlpha.MatchString(str)
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

// IsNumeric check if the string contains only numbers. Empty string is valid.
func IsNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxNumeric.MatchString(str)
}

// IsStringNumber is string number. should >= 0
func IsStringNumber(s string) bool {
	return s != "" && rxNumeric.MatchString(s)
}

// IsUTFNumeric check if the string contains only unicode numbers of any kind.
// Numbers can be 0-9 but also Fractions ¾,Roman Ⅸ and Hangzhou 〩. Empty string is valid.
func IsUTFNumeric(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}
	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if !unicode.IsNumber(c) { //numbers && minus sign are ok
			return false
		}
	}
	return true
}

// IsUTFDigit check if the string contains only unicode radix-10 decimal digits. Empty string is valid.
func IsUTFDigit(str string) bool {
	if IsNull(str) {
		return true
	}
	if strings.IndexAny(str, "+-") > 0 {
		return false
	}
	if len(str) > 1 {
		str = strings.TrimPrefix(str, "-")
		str = strings.TrimPrefix(str, "+")
	}
	for _, c := range str {
		if !unicode.IsDigit(c) { //digits && minus sign are ok
			return false
		}
	}
	return true

}

// IsEmail check if the string is an email.
func IsEmail(s string) bool {
	return s != "" && rxEmail.MatchString(s)
}

// IsExistingEmail check if the string is an email of existing domain
func IsExistingEmail(email string) bool {

	if len(email) < 6 || len(email) > 254 {
		return false
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return false
	}
	user := email[:at]
	host := email[at+1:]
	if len(user) > 64 {
		return false
	}
	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return false
	}
	switch host {
	case "localhost", "example.com":
		return true
	}
	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			return false
		}
	}

	return true
}

// IsByteLength check if the string's length (in bytes) falls in a range.
func IsByteLength(str string, min, max int) bool {
	return len(str) >= min && len(str) <= max
}

// IsUUIDv3 check if the string is a UUID version 3.
func IsUUIDv3(str string) bool {
	return rxUUID3.MatchString(str)
}

// IsUUIDv4 check if the string is a UUID version 4.
func IsUUIDv4(str string) bool {
	return rxUUID4.MatchString(str)
}

// IsUUIDv5 check if the string is a UUID version 5.
func IsUUIDv5(str string) bool {
	return rxUUID5.MatchString(str)
}

// IsUUID check if the string is a UUID (version 3, 4 or 5).
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

// IsCreditCard check if the string is a credit card.
func IsCreditCard(str string) bool {
	sanitized := notNumberRegexp.ReplaceAllString(str, "")
	if !rxCreditCard.MatchString(sanitized) {
		return false
	}
	var sum int64
	var digit string
	var tmpNum int64
	var shouldDouble bool
	for i := len(sanitized) - 1; i >= 0; i-- {
		digit = sanitized[i:(i + 1)]
		tmpNum, _ = ToInt(digit)
		if shouldDouble {
			tmpNum *= 2
			if tmpNum >= 10 {
				sum += ((tmpNum % 10) + 1)
			} else {
				sum += tmpNum
			}
		} else {
			sum += tmpNum
		}
		shouldDouble = !shouldDouble
	}

	return sum%10 == 0
}

// IsIP is the validation function for validating if the field's value is a valid v4 or v6 IP address.
func IsIP(s string) bool {
	// ip := net.ParseIP(s)
	return s != "" && net.ParseIP(s) != nil
}

// IsPort checks if a string represents a valid port
func IsPort(str string) bool {
	if i, err := strconv.Atoi(str); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsIPv4 is the validation function for validating if a value is a valid v4 IP address.
func IsIPv4(s string) bool {
	if s == "" {
		return false
	}

	ip := net.ParseIP(s)
	return ip != nil && strings.Contains(s, ".") // && ip.To4() != nil
}

// IsIPv6 is the validation function for validating if the field's value is a valid v6 IP address.
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && strings.Contains(s, ":") // && ip.To6() == nil
}

// IsMAC check if a string is valid MAC address.
// Possible MAC formats:
// 01:23:45:67:89:ab
// 01:23:45:67:89:ab:cd:ef
// 01-23-45-67-89-ab
// 01-23-45-67-89-ab-cd-ef
// 0123.4567.89ab
// 0123.4567.89ab.cdef
func IsMAC(s string) bool {
	if s == "" {
		return false
	}
	_, err := net.ParseMAC(s)
	return err == nil
}

// IsHost checks if the string is a valid IP (both v4 and v6) or a valid DNS name
func IsHost(str string) bool {
	return IsIP(str) || IsDNSName(str)
}

// IsMongoID check if the string is a valid hex-encoded representation of a MongoDB ObjectId.
func IsMongoID(str string) bool {
	return rxHexadecimal.MatchString(str) && (len(str) == 24)
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

// IsCIDR check if the string is an valid CIDR notiation (IPV4 & IPV6)
func IsCIDR(s string) bool {
	if s == "" {
		return false
	}

	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// IsJSON check if the string is valid JSON (note: uses json.Valid).
func IsJSON(s string) bool {
	if s == "" {
		return false
	}

	return json.ConfigCompatibleWithStandardLibrary.Valid(Bytes(s))
}

// HasLowerCase check if the string contains at least 1 lowercase. Empty string is valid.
func HasLowerCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasLowerCase.MatchString(str)
}

// HasUpperCase check if the string contains as least 1 uppercase. Empty string is valid.
func HasUpperCase(str string) bool {
	if IsNull(str) {
		return true
	}
	return rxHasUpperCase.MatchString(str)
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

// IsFilePath check is a string is Win or Unix file path and returns it's type.
func IsFilePath(str string) (bool, int) {
	if rxWinPath.MatchString(str) {
		//check windows path limit see:
		//  http://msdn.microsoft.com/en-us/library/aa365247(VS.85).aspx#maxpath
		if len(str[3:]) > 32767 {
			return false, Win
		}
		return true, Win
	} else if rxUnixPath.MatchString(str) {
		return true, Unix
	}
	return false, Unknown
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

// IsDivisibleBy check if the string is a number that's divisible by another.
// If second argument is not valid integer or zero, it's return false.
// Otherwise, if first argument is not valid integer or zero, it's return true (Invalid string converts to zero).
func IsDivisibleBy(str, num string) bool {
	f, _ := ToFloat(str)
	p := int64(f)
	q, _ := ToInt(num)
	if q == 0 {
		return false
	}
	return (p == 0) || (p%q == 0)
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("\\'\"!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

// IsSSN will validate the given string as a U.S. Social Security Number
func IsSSN(str string) bool {
	if str == "" || len(str) != 11 {
		return false
	}
	return rxSSN.MatchString(str)
}

// IsSemver check if string is valid semantic version
func IsSemver(str string) bool {
	return rxSemver.MatchString(str)
}

// IsType check if interface is of some type
func IsType(v interface{}, params ...string) bool {
	if len(params) == 1 {
		typ := params[0]
		return strings.Replace(reflect.TypeOf(v).String(), " ", "", -1) == strings.Replace(typ, " ", "", -1)
	}
	return false
}

// IsTime check if string is valid according to given format
func IsTime(str string, format string) bool {
	_, err := time.Parse(format, str)
	return err == nil
}

// IsUnixTime check if string is valid unix timestamp value
func IsUnixTime(str string) bool {
	if _, err := strconv.Atoi(str); err == nil {
		return true
	}
	return false
}

// IsRFC3339 check if string is valid timestamp value according to RFC3339
func IsRFC3339(str string) bool {
	return IsTime(str, time.RFC3339)
}

// IsRFC3339WithoutZone check if string is valid timestamp value according to RFC3339 which excludes the timezone.
func IsRFC3339WithoutZone(str string) bool {
	return IsTime(str, RF3339WithoutZone)
}

// IsISO4217 check if string is valid ISO currency code
func IsISO4217(str string) bool {
	for _, currency := range ISO4217List {
		if str == currency {
			return true
		}
	}

	return false
}

// IsNull check if the string is null.
func IsNull(str string) bool {
	return len(str) == 0
}

// IsNotNull check if the string is not null.
func IsNotNull(str string) bool {
	return !IsNull(str)
}

// HasWhitespaceOnly checks the string only contains whitespace
func HasWhitespaceOnly(str string) bool {
	return len(str) > 0 && rxHasWhitespaceOnly.MatchString(str)
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

// ByteLength check string's length, minLen, maxLen.
func ByteLength(str string, params ...string) bool {
	if len(params) == 2 {
		min, _ := ToInt(params[0])
		max, _ := ToInt(params[1])
		return len(str) >= int(min) && len(str) <= int(max)
	}

	return false
}

// RuneLength check string's length, minLen, maxLen.
// Alias for StringLength
func RuneLength(str string, params ...string) bool {
	return StringLength(str, params...)
}

// RuneLength check string's length (including multi byte strings)
func RuneLength2(val interface{}, minLen int, maxLen ...int) bool {
	str, isString := val.(string)
	if !isString {
		return false
	}

	strLen := utf8.RuneCountInString(str)

	// only min length check.
	if len(maxLen) == 0 {
		return strLen >= minLen
	}

	// min and max length check
	return strLen >= minLen && strLen <= maxLen[0]
}

// IsRsaPub check whether string is valid RSA key
// Alias for IsRsaPublicKey
func IsRsaPub(str string, params ...string) bool {
	if len(params) == 1 {
		len, _ := ToInt(params[0])
		return IsRsaPublicKey(str, int(len))
	}

	return false
}

// StringMatches checks if a string matches a given pattern.
func StringMatches(s string, params ...string) bool {
	if len(params) == 1 {
		pattern := params[0]
		return Matches(s, pattern)
	}
	return false
}

// StringLength check string's length (including multi byte strings)
func StringLength(str string, params ...string) bool {
	if len(params) == 2 {
		strLength := utf8.RuneCountInString(str)
		min, _ := ToInt(params[0])
		max, _ := ToInt(params[1])
		return strLength >= int(min) && strLength <= int(max)
	}

	return false
}

// StringLength check string's length (including multi byte strings)
func StringLength2(val interface{}, minLen int, maxLen ...int) bool {
	return RuneLength2(val, minLen, maxLen...)
}

// MinStringLength check string's minimum length (including multi byte strings)
func MinStringLength(str string, params ...string) bool {

	if len(params) == 1 {
		strLength := utf8.RuneCountInString(str)
		min, _ := ToInt(params[0])
		return strLength >= int(min)
	}

	return false
}

// MaxStringLength check string's maximum length (including multi byte strings)
func MaxStringLength(str string, params ...string) bool {

	if len(params) == 1 {
		strLength := utf8.RuneCountInString(str)
		max, _ := ToInt(params[0])
		return strLength <= int(max)
	}

	return false
}

// Range check string's length
func Range(str string, params ...string) bool {
	if len(params) == 2 {
		value, _ := ToFloat(str)
		min, _ := ToFloat(params[0])
		max, _ := ToFloat(params[1])
		return InRange(value, min, max)
	}

	return false
}

func IsInRaw(str string, params ...string) bool {
	if len(params) == 1 {
		rawParams := params[0]

		parsedParams := strings.Split(rawParams, "|")

		return IsIn(str, parsedParams...)
	}

	return false
}

// IsIn check if string str is a member of the set of strings params
func IsIn(str string, params ...string) bool {
	for _, param := range params {
		if str == param {
			return true
		}
	}

	return false
}
