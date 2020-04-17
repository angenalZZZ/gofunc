package random

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestAlphaNumber(t *testing.T) {
	for i := range f.TimesRepeat(10, 1) {
		s := AlphaNumberLower(i + 1)
		t.Log(s)
	}
}
