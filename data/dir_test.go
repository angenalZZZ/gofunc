package data

import (
	"testing"

	"github.com/angenalZZZ/gofunc/f"
)

func TestDir(t *testing.T) {
	t.Log(RootDir, f.PathExists(RootDir))
	t.Log(CurrentPath)
	t.Log(CurrentDir, CurrentFile)
	t.Log(CurrentUserName, CurrentUserHomeDir, f.PathExists(CurrentUserHomeDir))
}
