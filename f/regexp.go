package f

import "regexp"

// validator.go
// Basic regular expressions for validating strings.

const (
	RxTagName           string = "valid"
	RxEmail             string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	RxCreditCard        string = "^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|(222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11}|6[27][0-9]{14})$"
	RxISBN10            string = "^(?:[0-9]{9}X|[0-9]{10})$"
	RxISBN13            string = "^(?:[0-9]{13})$"
	RxUUID3             string = "^[0-9a-f]{8}-[0-9a-f]{4}-3[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$"
	RxUUID4             string = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	RxUUID5             string = "^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
	RxUUID              string = "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	RxAlpha             string = "^[a-zA-Z]+$"
	RxAlphaNumeric      string = "^[a-zA-Z0-9]+$"
	RxAlphaDash         string = `^(?:[\w-]+)$`
	RxNumeric           string = "^[0-9]+$"
	RxInt               string = "^(?:[-+]?(?:0|[1-9][0-9]*))$"
	RxFloat             string = "^(?:[-+]?(?:[0-9]+))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$"
	RxCnMobile          string = `^1\d{10}$`
	RxHexadecimal       string = "^[0-9a-fA-F]+$"
	RxHexColor          string = "^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$"
	RxRGBColor          string = "^rgb\\(\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*,\\s*(0|[1-9]\\d?|1\\d\\d?|2[0-4]\\d|25[0-5])\\s*\\)$"
	RxASCII             string = "^[\x00-\x7F]+$"
	RxMultiByte         string = "[^\x00-\x7F]"
	RxFullWidth         string = "[^\u0020-\u007E\uFF61-\uFF9F\uFFA0-\uFFDC\uFFE8-\uFFEE0-9a-zA-Z]"
	RxHalfWidth         string = "[\u0020-\u007E\uFF61-\uFF9F\uFFA0-\uFFDC\uFFE8-\uFFEE0-9a-zA-Z]"
	RxBase64            string = "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	RxPrintableASCII    string = "^[\x20-\x7E]+$"
	RxDataURI           string = "^data:.+\\/(.+);base64$" // `^data:.+/(.+);base64,(?:.+)`
	RxMagnetURI         string = "^magnet:\\?xt=urn:[a-zA-Z0-9]+:[a-zA-Z0-9]{32,40}&dn=.+&tr=.+$"
	RxLatitude          string = "^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$"
	RxLongitude         string = "^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$"
	RxDNSName           string = `^([a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62}){1}(\.[a-zA-Z0-9_]{1}[a-zA-Z0-9_-]{0,62})*[\._]?$`
	RxIP                string = `(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`
	RxFullURL           string = `^(?:ftp|tcp|udp|wss?|https?):\/\/[\w\.\/#=?&]+$`
	RxURLSchema         string = `((ftp|tcp|udp|wss?|https?):\/\/)`
	RxURLUsername       string = `(\S+(:\S*)?@)`
	RxURLPath           string = `((\/|\?|#)[^\s]*)`
	RxURLPort           string = `(:(\d{1,5}))`
	RxUrlIP             string = `([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))`
	RxURLSubDomain      string = `((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))`
	RxURL               string = `^` + RxURLSchema + `?` + RxURLUsername + `?` + `((` + RxUrlIP + `|(\[` + RxIP + `\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(` + RxURLSubDomain + `?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?` + RxURLPort + `?` + RxURLPath + `?$`
	RxSSN               string = `^\d{3}[- ]?\d{2}[- ]?\d{4}$`
	RxWinPath           string = `^[a-zA-Z]:\\(?:[^\\/:*?"<>|\r\n]+\\)*[^\\/:*?"<>|\r\n]*$`
	RxUnixPath          string = `^(/[^/\x00]*)+/?$`
	RxSemver            string = "^v?(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)\\.(?:0|[1-9]\\d*)(-(0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(\\.(0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\\+[0-9a-zA-Z-]+(\\.[0-9a-zA-Z-]+)*)?$"
	RxHasLowerCase      string = ".*[[:lower:]]"
	RxHasUpperCase      string = ".*[[:upper:]]"
	RxHasWhitespace     string = ".*[[:space:]]"
	RxHasWhitespaceOnly string = "^[[:space:]]+$"
)

var (
	userRegexp          = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp          = regexp.MustCompile("^[^\\s]+\\.[^\\s]+$")
	userDotRegexp       = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
	rxEmail             = regexp.MustCompile(RxEmail)
	rxCreditCard        = regexp.MustCompile(RxCreditCard)
	rxISBN10            = regexp.MustCompile(RxISBN10)
	rxISBN13            = regexp.MustCompile(RxISBN13)
	rxUUID3             = regexp.MustCompile(RxUUID3)
	rxUUID4             = regexp.MustCompile(RxUUID4)
	rxUUID5             = regexp.MustCompile(RxUUID5)
	rxUUID              = regexp.MustCompile(RxUUID)
	rxAlpha             = regexp.MustCompile(RxAlpha)
	rxAlphaNumeric      = regexp.MustCompile(RxAlphaNumeric)
	rxAlphaDash         = regexp.MustCompile(RxAlphaDash)
	rxNumeric           = regexp.MustCompile(RxNumeric)
	rxInt               = regexp.MustCompile(RxInt)
	rxFloat             = regexp.MustCompile(RxFloat)
	rxCnMobile          = regexp.MustCompile(RxCnMobile)
	rxHexadecimal       = regexp.MustCompile(RxHexadecimal)
	rxHexColor          = regexp.MustCompile(RxHexColor)
	rxRGBColor          = regexp.MustCompile(RxRGBColor)
	rxASCII             = regexp.MustCompile(RxASCII)
	rxPrintableASCII    = regexp.MustCompile(RxPrintableASCII)
	rxMultiByte         = regexp.MustCompile(RxMultiByte)
	rxFullWidth         = regexp.MustCompile(RxFullWidth)
	rxHalfWidth         = regexp.MustCompile(RxHalfWidth)
	rxBase64            = regexp.MustCompile(RxBase64)
	rxDataURI           = regexp.MustCompile(RxDataURI)
	rxMagnetURI         = regexp.MustCompile(RxMagnetURI)
	rxLatitude          = regexp.MustCompile(RxLatitude)
	rxLongitude         = regexp.MustCompile(RxLongitude)
	rxDNSName           = regexp.MustCompile(RxDNSName)
	rxFullURL           = regexp.MustCompile(RxFullURL)
	rxURLSchema         = regexp.MustCompile(RxURLSchema)
	rxURL               = regexp.MustCompile(RxURL)
	rxSSN               = regexp.MustCompile(RxSSN)
	rxWinPath           = regexp.MustCompile(RxWinPath)
	rxUnixPath          = regexp.MustCompile(RxUnixPath)
	rxSemver            = regexp.MustCompile(RxSemver)
	rxHasLowerCase      = regexp.MustCompile(RxHasLowerCase)
	rxHasUpperCase      = regexp.MustCompile(RxHasUpperCase)
	rxHasWhitespace     = regexp.MustCompile(RxHasWhitespace)
	rxHasWhitespaceOnly = regexp.MustCompile(RxHasWhitespaceOnly)
)
