package random

import "github.com/dgraph-io/ristretto/z"

var Int = z.FastRand

func Int8() uint8 {
	return uint8(z.FastRand() % (1 << 8))
}

func Int16() uint16 {
	return uint16(z.FastRand() % (1 << 16))
}

func Max(val1, val2 uint32) uint32 {
	if val1 == val2 {
		return val1
	}

	if val1 > val2 {
		val1, val2 = val2, val1
	}

	return val1 + z.FastRand()%(val2-val1)
}
