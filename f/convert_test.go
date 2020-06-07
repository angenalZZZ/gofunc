package f

import "testing"

func TestToInt(t *testing.T) {
	// ToInt parse string to int64
	if i, err := ToInt("123", false); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("ToInt: %d\n", i)
	}

	raw := "hello world"
	rawData := []byte(raw)
	// Int32Sum returns the CRC-32 checksum of data using the IEEE polynomial.
	int32x := Int32Sum(rawData)
	t.Logf("Int32Sum: %d\n", int32x)
	int32x = Int32SumString(raw)
	t.Logf("Int32SumString: %d\n", int32x)
	// Int64Sum xx.hash is a Go implementation of the 64-bit xxHash algorithm, XXH64.
	int64x := Int64Sum(rawData)
	t.Logf("Int64Sum: %d\n", int64x)
	int64x = Int64SumString(raw)
	t.Logf("Int64SumString: %d\n", int64x)
	int64x = Int64MemHash(rawData)
	t.Logf("Int64MemHash: %d\n", int64x)
	int64x = Int64MemHashString(raw)
	t.Logf("Int64MemHashString: %d\n", int64x)

	buf1 := BytesUint16(65535)
	i16 := Uint16Bytes(buf1)
	if i16 != 1<<16-1 {
		t.Fatal()
	}

	buf2 := BytesUint32(4294967295)
	i32 := Uint32Bytes(buf2)
	if i32 != 1<<32-1 {
		t.Fatal()
	}

	buf3 := BytesUint64(18446744073709551615)
	i64 := Uint64Bytes(buf3)
	if i64 != 1<<64-1 {
		t.Fatal()
	}
}
