package data

import (
	"github.com/angenalZZZ/gofunc/f"
	"sync"
)

var (
	RootDir            = `A:\test`
	CurrentDir         = f.CurrentDir()
	CurrentPath        = f.CurrentPath()
	CurrentUserName    = f.CurrentUserName()
	CurrentUserHomeDir = f.CurrentUserHomeDir()

	Init = new(sync.Once).Do
)
