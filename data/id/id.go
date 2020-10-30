package id

import (
	"github.com/angenalZZZ/gofunc/data/id/shortid"
	"github.com/rs/xid"
)

func L10() string {
	s := shortid.MustGenerate()
	if len(s) == 10 {
		return s
	}
	return s + "0"
}

func L20() string {
	s := xid.New().String()
	return s
}
