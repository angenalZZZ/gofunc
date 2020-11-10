package data

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/angenalZZZ/gofunc/f"
)

func TestDir(t *testing.T) {
	t.Log(RootDir, f.PathExists(RootDir))
	t.Log(CurrentPath)
	t.Log(CurrentDir, CurrentFile)
	t.Log(CurrentUserName, CurrentUserHomeDir, f.PathExists(CurrentUserHomeDir))

	path := filepath.Join(CurrentDir, ".nats")
	if exists, isDir := f.FileExist(path); !exists || !isDir {
		if exists {
			_ = os.RemoveAll(path)
		}
		if err := os.Mkdir(path, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Log(path, f.PathExists(path), f.IsDir(path))
}
