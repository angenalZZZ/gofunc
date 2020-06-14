package main

import (
	"flag"
	"fmt"
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// go get github.com/angenalZZZ/gofunc/data/cache/fastcache/cmd/fastcache1e6
// go build -ldflags "-s -w" -o A:/test/ .
// cd A:/test/ && fastcache1e6 -c 2 -d 128 -t 10000000

var (
	flagCont   = flag.Int("c", 1, "total threads")
	flagData   = flag.Int("d", 128, "every time request bytes")
	flagTimes  = flag.Int("t", 1000000, "total times")
	flagRemove = flag.Bool("r", true, "delete data files")
)

func init() {
	flag.Usage = func() {
		fmt.Printf(" Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	if len(os.Args) < 3 {
		flag.Usage()
		return
	}

	fmt.Println()
	c, l, m := *flagCont, *flagData, *flagTimes

	p := []byte(random.AlphaNumberLower(l))
	tl := fastcache.NewTimeline(time.Now(), time.Now().Add(time.Hour), time.Hour, f.CurrentDir(), 2048)

	wg := new(sync.WaitGroup)
	wg.Add(c)
	t1 := time.Now()
	for x := 0; x < c; x++ {
		n := m / (x + 1)
		go func(n int) {
			for i := 0; i < n; i++ {
				_, _ = tl.Write(p)
			}
			wg.Done()
		}(n)
	}
	wg.Wait()
	t2 := time.Now()
	tl.Save()

	fmt.Printf(" every time %d bytes data request \n", l)
	fmt.Printf(" take requested time %s \n", t2.Sub(t1))
	fmt.Printf(" take saved time %s \n", time.Now().Sub(t2))
	s, e := ioutil.ReadFile(filepath.Join(tl.CacheDir, tl.Frames[0].Filename()))
	fmt.Printf(" %s %v \n", s, e)
	if *flagRemove {
		tl.RemoveAll()
	}
}
