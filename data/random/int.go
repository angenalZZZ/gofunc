package random

import "github.com/angenalZZZ/gofunc/g"

var Int = g.FastRand

func Int8() uint8 {
	return uint8(g.FastRand() % (1 << 8))
}

func Int16() uint16 {
	return uint16(g.FastRand() % (1 << 16))
}

func Max(val1, val2 uint32) uint32 {
	if val1 == val2 {
		return val1
	}

	if val1 > val2 {
		val1, val2 = val2, val1
	}

	return val1 + g.FastRand()%(val2-val1)
}
