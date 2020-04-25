package random

import "testing"

func TestInt(t *testing.T) {
	t.Log(Int())
	t.Log(Int8())
	t.Log(Int16())
	t.Log(Max(1000, 10000))
	t.Log(Max(100, 1000))
	t.Log(Max(10, 100))
	t.Log(Max(0, 10))
}
