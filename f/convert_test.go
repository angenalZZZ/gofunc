package f

import "testing"

func TestToInt(t *testing.T) {
	// ToInt parse string to int64
	if i, err := ToInt("123", false); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("ToInt: %d\n", i)
	}

	// Int32Sum returns the CRC-32 checksum of data using the IEEE polynomial.
	int32x := Int32Sum([]byte("hello world"))
	t.Logf("Int32Sum: %d\n", int32x)
	int32x = Int32SumString("hello world")
	t.Logf("Int32SumString: %d\n", int32x)
	// Int64Sum xx.hash is a Go implementation of the 64-bit xxHash algorithm, XXH64.
	int64x := Int64Sum([]byte("hello world"))
	t.Logf("Int64Sum: %d\n", int64x)
	int64x = Int64SumString("hello world")
	t.Logf("Int64SumString: %d\n", int64x)
}
