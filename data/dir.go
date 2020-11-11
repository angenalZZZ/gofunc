package data

import (
	"path/filepath"
	"sync"

	"github.com/angenalZZZ/gofunc/f"
)

var (
	// RootDir todo sets application root dir
	RootDir = `A:\test`
	// CurrentDir gets current dir
	CurrentDir = f.CurrentDir()
	// CurrentPath gets current path
	CurrentPath = f.CurrentPath()
	// CurrentFile gets current file name
	CurrentFile = f.CurrentFile()
	// CurrentUserName gets computer name - user name
	CurrentUserName = f.CurrentUserName()
	// CurrentUserHomeDir get $HOME
	CurrentUserHomeDir = f.CurrentUserHomeDir()

	// Dir get the new directory under the current directory
	Dir = func(name string) string { return filepath.Join(CurrentDir, name) }

	// Init todo init function
	Init = new(sync.Once).Do
)
