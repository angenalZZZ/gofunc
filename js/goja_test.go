package js

import (
	"testing"

	"github.com/dop251/goja"
)

func TestConsole(t *testing.T) {
	r := goja.New()
	Console(r)
	if v, err := r.RunString(`console.log('hello world,', new Date)`); err != nil {
		t.Fatal(err)
	} else if !v.Equals(goja.Undefined()) {
		t.Fail()
	}
}
