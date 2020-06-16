package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestRound(t *testing.T) {
	v := 0.22673826
	f1, f2 := f.Round(v, 3), f.Floor(v, 5)
	// Output: 0.227
	t.Log(f1)
	// Output: 0.22673
	t.Log(f2)
}
