package f

import "os"

// ExistsFile Exists File.
func ExistsFile(file string) bool {
	if _, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// ExistsDir Exists Dir.
func ExistsDir(file string) bool {
	if stat, err := os.Stat(file); err != nil && os.IsNotExist(err) {
		return false
	} else {
		return stat.IsDir()
	}
}
