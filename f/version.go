package f

import (
	"runtime"
	"strings"
)

var GoVersion = strings.TrimPrefix(runtime.Version(), "go")
