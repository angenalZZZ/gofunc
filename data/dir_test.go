package data_test

import (
	"runtime"
	"testing"

	"github.com/angenalZZZ/gofunc/data"
	"github.com/angenalZZZ/gofunc/f"
)

func init() {
	data.Init(func() {
		_ = f.MkdirCurrent(".nats")
		_ = f.MkdirCurrent(".nutsdb")
	})
}

func TestDir(t *testing.T) {
	t.Log(runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), "cpu")
	t.Logf("cmd dir %s", data.CodeDir("cmd"))
	t.Logf("data dir %s", data.CodeDirname)

	println()
	t.Log(data.RootDir, f.PathExists(data.RootDir))
	t.Log(data.CurrentPath)
	t.Log(data.CurrentDir, data.CurrentFile)
	t.Log(data.CurrentUserName, data.CurrentUserHomeDir, f.PathExists(data.CurrentUserHomeDir))

	println()
	path := data.Dir(".nats")
	t.Log(path, "--Mkdir--", f.IsDir(path))
	path = data.Dir(".nutsdb")
	t.Log(path, "--Mkdir--", f.IsDir(path))
}
