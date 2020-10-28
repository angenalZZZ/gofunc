package f_test

import (
	"bytes"
	"testing"

	"github.com/angenalZZZ/gofunc/f"
)

func TestGbkToUtf8(t *testing.T) {
	data, err := f.ReadFile("../test/temp/encoding-gbk.txt")
	if err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%s", data)
	}
	if data, err = f.GbkToUtf8(data); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("%s", data)
	}
}

func TestToInt(t *testing.T) {
	// ToInt parse string to int64
	if i, err := f.ToInt("123"); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("ToInt: %d\n", i)
	}

	raw := "hello world"
	rawData := []byte(raw)
	// Int32Sum returns the CRC-32 checksum of data using the IEEE polynomial.
	int32x := f.Int32Sum(rawData)
	t.Logf("Int32Sum: %d\n", int32x)
	int32x = f.Int32SumString(raw)
	t.Logf("Int32SumString: %d\n", int32x)
	// Int64Sum xx.hash is a Go implementation of the 64-bit xxHash algorithm, XXH64.
	int64x := f.Int64Sum(rawData)
	t.Logf("Int64Sum: %d\n", int64x)
	int64x = f.Int64SumString(raw)
	t.Logf("Int64SumString: %d\n", int64x)
	int64x = f.Int64MemHash(rawData)
	t.Logf("Int64MemHash: %d\n", int64x)
	int64x = f.Int64MemHashString(raw)
	t.Logf("Int64MemHashString: %d\n", int64x)

	buf1 := f.BytesUint16(65535)
	i16 := f.Uint16Bytes(buf1)
	if i16 != 1<<16-1 {
		t.Fatal()
	}

	buf2 := f.BytesUint32(4294967295)
	i32 := f.Uint32Bytes(buf2)
	if i32 != 1<<32-1 {
		t.Fatal()
	}

	buf3 := f.BytesUint64(18446744073709551615)
	i64 := f.Uint64Bytes(buf3)
	if i64 != 1<<64-1 {
		t.Fatal()
	}
}

func TestMapMerge(t *testing.T) {
	for _, tuple := range []struct {
		src      string
		dst      string
		expected string
	}{
		{
			src:      `{}`,
			dst:      `{}`,
			expected: `{}`,
		},
		{
			src:      `{"b":2}`,
			dst:      `{"a":1}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			src:      `{"a":0}`,
			dst:      `{"a":1}`,
			expected: `{"a":0}`,
		},
		{
			src:      `{"a":{       "y":2}}`,
			dst:      `{"a":{"x":1       }}`,
			expected: `{"a":{"x":1, "y":2}}`,
		},
		{
			src:      `{"a":{"x":2}}`,
			dst:      `{"a":{"x":1}}`,
			expected: `{"a":{"x":2}}`,
		},
		{
			src:      `{"a":{       "y":7, "z":8}}`,
			dst:      `{"a":{"x":1, "y":2       }}`,
			expected: `{"a":{"x":1, "y":7, "z":8}}`,
		},
		{
			src:      `{"1": { "b":1, "2": { "3": {         "b":3, "n":[1,2]} }        }}`,
			dst:      `{"1": {        "2": { "3": {"a":"A",        "n":"xxx"} }, "a":3 }}`,
			expected: `{"1": { "b":1, "2": { "3": {"a":"A", "b":3, "n":[1,2]} }, "a":3 }}`,
		},
	} {
		var dst map[string]interface{}
		if err := f.DecodeJson([]byte(tuple.dst), &dst); err != nil {
			t.Error(err)
			continue
		}

		var src map[string]interface{}
		if err := f.DecodeJson([]byte(tuple.src), &src); err != nil {
			t.Error(err)
			continue
		}

		var expected map[string]interface{}
		if err := f.DecodeJson([]byte(tuple.expected), &expected); err != nil {
			t.Error(err)
			continue
		}

		got := f.MapMerge(dst, src)
		assertMapMerge(t, expected, got)
	}
}

func assertMapMerge(t *testing.T, expected, got map[string]interface{}) {
	expectedBuf, err := f.EncodeJson(expected)
	if err != nil {
		t.Error(err)
		return
	}
	gotBuf, err := f.EncodeJson(got)
	if err != nil {
		t.Error(err)
		return
	}
	if bytes.Compare(expectedBuf, gotBuf) != 0 {
		t.Errorf("expected %s, got %s", string(expectedBuf), string(gotBuf))
		return
	}
}
