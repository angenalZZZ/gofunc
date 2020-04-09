package f

import "testing"

func TestToInt(t *testing.T) {
	// ToInt parse string to int64
	if i, err := ToInt("123", false); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("ToInt: %d\n", i)
	}

	// ToIntCrc32IEEE returns the CRC-32 checksum of data using the IEEE polynomial.
	i := ToIntCrc32IEEE([]byte("hello world"))
	t.Logf("ToIntCrc32IEEE: %d\n", i)
}
