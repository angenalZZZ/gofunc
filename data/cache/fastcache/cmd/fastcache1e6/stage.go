package main

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/data/cache/fastcache"
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

func Stage() {
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
