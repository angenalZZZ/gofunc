package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestGoVersion(t *testing.T) {
	t.Logf("Go Version: %s", f.GoVersion)
}

func TestNumCPU(t *testing.T) {
	t.Logf("Number of CPUs: %d * 16 = %d", f.NumCPU, f.NumCPUx16)
}
