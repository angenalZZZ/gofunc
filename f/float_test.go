package f

import "testing"

func TestRound(t *testing.T) {
	v := 0.22673826
	f1, f2 := Round(v, 3), Floor(v, 5)
	// Output: 0.227
	t.Log(f1)
	// Output: 0.22673
	t.Log(f2)
}
