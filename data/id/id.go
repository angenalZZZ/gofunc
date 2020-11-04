package id

import (
	"github.com/angenalZZZ/gofunc/data/id/shortid"
	"github.com/google/uuid"
	"github.com/rs/xid"
)

// L9 short id 9 chars
func L9() string {
	s := shortid.MustGenerate()
	if len(s) == 10 {
		return s[0:9]
	}
	return s
}

// L10 short id 10 chars
func L10() string {
	s := shortid.MustGenerate()
	if len(s) == 10 {
		return s
	}
	return s + "0"
}

// L20 xid 20 chars
func L20() string {
	s := xid.New().String()
	return s
}

// L32 uuid 32 chars
func L32() string {
	s := uuid.New().String()
	return s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
}

// L36 uuid 36 chars
func L36() string {
	s := uuid.New().String()
	return s
}
