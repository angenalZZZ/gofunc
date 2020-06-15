package f

import (
	"reflect"
	"runtime"
	"strings"
)

const (
	MimeSniffLen = 512 // sniff Length, use for detect file mime type
)

// Used by IsFilePath func
const (
	// Unknown is unresolved OS type
	Unknown = iota
	// Win is Windows type
	Win
	// Unix is *nix OS types
	Unix
)

// time_stamp.go
// TimeFormat https://programming.guide/go/format-parse-string-time-date-example.html
const (
	DateFormatStringG     string = "20060102"
	TimeFormatStringS     string = "20060102150405"
	TimeFormatStringM     string = "20060102150405.000"
	DateFormatString      string = "2006-01-02"
	DateTimeFormatStringH string = "2006-01-02 15"
	DateTimeFormatStringM string = "2006-01-02 15:04"
	DateTimeFormatString  string = "2006-01-02 15:04:05"
	TimeFormatString      string = "2006-01-02 15:04:05.000"
)

var (
	emptyValue = reflect.Value{}

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
