package f

import (
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

const (
	MimeSniffLen = 512 // sniff Length, use for detect file mime type
)

// time_stamp.go
// TimeFormat https://programming.guide/go/format-parse-string-time-date-example.html
const (
	DateFormatStringG     string = "20060102"
	DateFormatString      string = "2006-01-02"
	DateTimeFormatStringN string = "20060102150405"
	DateTimeFormatStringH string = "2006-01-02 15"
	DateTimeFormatStringM string = "2006-01-02 15:04"
	DateTimeFormatString  string = "2006-01-02 15:04:05"
	TimeFormatString      string = "2006-01-02 15:04:05.000"
)

var (
	// Go Version Numbers
	GoVersion = strings.TrimPrefix(runtime.Version(), "go")

	// IsImageFile: refer net/http package
	ImageMimeTypes = map[string]string{
		"bmp":  "image/bmp",
		"gif":  "image/gif",
		"ief":  "image/ief",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"svg":  "image/svg+xml",
		"ico":  "image/x-icon",
		"webp": "image/webp",
	}
)

// validator.go
// Basic regular expressions for validating strings.
var (
	emptyValue       = reflect.Value{}
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

// zip.go
// CompressedFormats is a (non-exhaustive) set of lowerCased
// file extensions for formats that are typically already
// compressed. Compressing files that are already compressed
// is inefficient, so use this set of extension to avoid that.
var CompressedFormats = map[string]struct{}{
	".7z":   {},
	".avi":  {},
	".br":   {},
	".bz2":  {},
	".cab":  {},
	".docx": {},
	".gif":  {},
	".gz":   {},
	".jar":  {},
	".jpeg": {},
	".jpg":  {},
	".lz":   {},
	".lz4":  {},
	".lzma": {},
	".m4v":  {},
	".mov":  {},
	".mp3":  {},
	".mp4":  {},
	".mpeg": {},
	".mpg":  {},
	".png":  {},
	".pptx": {},
	".rar":  {},
	".sz":   {},
	".tbz2": {},
	".tgz":  {},
	".tsz":  {},
	".txz":  {},
	".xlsx": {},
	".xz":   {},
	".zip":  {},
	".zipx": {},
}
