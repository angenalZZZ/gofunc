package data

import (
	"path/filepath"
	"runtime"
	"sync"

	"github.com/angenalZZZ/gofunc/f"
)

var (
	// RootDir todo sets application root dir
	RootDir = `A:\test`
	// CurrentDir get current dir
	CurrentDir = f.CurrentDir()
	// CurrentPath get current path
	CurrentPath = f.CurrentPath()
	// CurrentFile get current file name
	CurrentFile = f.CurrentFile()
	// CurrentUserName get computer name - user name
	CurrentUserName = f.CurrentUserName()
	// CurrentUserHomeDir get $HOME
	CurrentUserHomeDir = f.CurrentUserHomeDir()
	// CodeDir get the directory where the current code file is located
	CodeDirname, _ = CodeDirFileLine()

	// CodeDir get current code dir
	CodeDir = func(codeDir string) string { return filepath.Join(filepath.Dir(CodeDirname), codeDir) }

	// CodeDirFileLine get current code file name and line number
	CodeDirFileLine = func() (string, int) {
		_, file, line, _ := runtime.Caller(0)
		return filepath.Dir(file), line
	}

	// Dir get the new directory under the current directory
	Dir = func(name string) string { return filepath.Join(CurrentDir, name) }

	// Init todo init function
	Init = new(sync.Once).Do
)
