package data

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestDir(t *testing.T) {
	t.Log(RootDir, f.PathExists(RootDir))
	t.Log(CurrentDir, f.PathExists(CurrentDir))
	t.Log(CurrentPath, f.PathExists(CurrentPath))
	t.Log(CurrentUserName, CurrentUserHomeDir, f.PathExists(CurrentUserHomeDir))
}
