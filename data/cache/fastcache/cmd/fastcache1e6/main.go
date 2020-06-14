package main

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"os"
	"time"
)

// go get github.com/angenalZZZ/gofunc/data/cache/fastcache/cmd/fastcache1e6
func main() {
	l := 128 // every time 128B data request
	if len(os.Args) > 1 && f.IsInt(os.Args[1]) {
		n, _ := f.ToInt(os.Args[1], false)
		if n > 0 {
			l = int(n)
		}
	}

	m := 1000000 // request times
	if len(os.Args) > 2 && f.IsInt(os.Args[2]) {
		n, _ := f.ToInt(os.Args[2], false)
		if n > 0 {
			m = int(n)
		}
	}

	fmt.Println()
	tl := fastcache.NewTimeline(time.Now(), time.Now().Add(time.Hour), time.Hour, f.CurrentDir(), 2048)
	p := []byte(random.AlphaNumberLower(l))

	t1 := time.Now()
	for i := 0; i < m; i++ {
		_, _ = tl.Write(p)
	}

	t2 := time.Now()
	tl.Save()

	fmt.Printf(" every time %d bytes data request, total %d times \n", l, m)
	fmt.Printf(" take requested time %s \n", t2.Sub(t1))
	fmt.Printf(" take saved time %s \n\n", time.Now().Sub(t2))
	tl.RemoveAll()
}
