package random

import (
	"math/rand"
	"os"
	"time"
)

var (
	Numbers   = "0123456789"
	LowerCase = "abcdefghijklmnopqrstuvwxyz"
	UpperCase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//SpecialChars = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

	AlphaNumbers      = Numbers + LowerCase + UpperCase
	AlphaNumbersLower = Numbers + LowerCase
	AlphaNumbersUpper = Numbers + UpperCase
)

// R generator
var R = rand.New(rand.NewSource(time.Now().UnixNano() + int64(os.Getpid())))

func Number(length int) string           { return String(Numbers, length) }
func AlphaNumber(length int) string      { return String(AlphaNumbers, length) }
func AlphaNumberLower(length int) string { return String(AlphaNumbersLower, length) }
func AlphaNumberUpper(length int) string { return String(AlphaNumbersUpper, length) }

func NumberBytes(length int) []byte           { return Bytes(Numbers, length) }
func AlphaNumberBytes(length int) []byte      { return Bytes(AlphaNumbers, length) }
func AlphaNumberLowerBytes(length int) []byte { return Bytes(AlphaNumbersLower, length) }
func AlphaNumberUpperBytes(length int) []byte { return Bytes(AlphaNumbersUpper, length) }

func Bytes(chooseFrom string, length int) []byte {
	l := len(chooseFrom)
	p := make([]byte, length)
	for i := range p {
		p[i] = chooseFrom[R.Intn(l)]
	}
	return p
}

func String(chooseFrom string, length int) string {
	return string(Bytes(chooseFrom, length))
}
