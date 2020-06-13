package fastcache

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// Timeline time data
type Timeline struct {
	CacheDir string // cache persist to disk directory
	frames   []*timeFrame
	duration time.Duration
	index    int64
}

// timeFrame time bounds on which data to retrieve.
type timeFrame struct {
	cache *Cache // a fast thread-safe inmemory cache optimized for big number of entries.
	frame *f.TimeFrame
	index uint32
}

func (t *Timeline) Write(p []byte) (n int, err error) {
	if t.index == -1 {
		return
	}

	c := t.frames[t.index]
	i := atomic.AddUint32(&c.index, 1)
	c.cache.Set(f.BytesUint32(i), p)
	return int(i), nil
}

func (c *timeFrame) dirname() string {
	return fmt.Sprintf("%s.%d", c.frame.Since.LocalTimeStampString(true), c.index)
}

func (c *timeFrame) save(cacheDir string) {
	time.Sleep(time.Microsecond)
	if c.index == 0 {
		return
	}

	fileStat := new(Stats)
	c.cache.UpdateStats(fileStat)
	data, err := f.EncodeJson(fileStat)
	logErr := log.New(os.Stderr, "", 0)
	if err != nil {
		logErr.Print(err)
	}

	filePath := filepath.Join(cacheDir, c.dirname())
	err = ioutil.WriteFile(filePath+".json", data, 0644)
	if err != nil {
		logErr.Print(err)
	}

	if err = c.cache.SaveToFileConcurrent(filePath, 0); err != nil {
		logErr.Print(err)
	} else {
		c.cache.Reset() // Reset removes all the items from the cache.
	}
}

func (t *Timeline) init() {
	p := int64(t.duration.Seconds())
	n := t.frames[0].frame.Since.UnixSecond
	m := t.frames[len(t.frames)-1].frame.Until.UnixSecond
	for u := time.Now().Unix(); u < m; u++ {
		index := (u - n) / p
		if index >= 0 && index != t.index {
			if t.index != -1 {
				go t.frames[t.index].save(t.CacheDir)
			}
			atomic.StoreInt64(&t.index, index)
		}
		time.Sleep(time.Second)
	}
	t.index = -1
}

func NewTimeline(since, until time.Time, duration time.Duration, cacheDir string, maxBytes int) *Timeline {
	frames := f.NewTimeFrames(since, until, duration)

	t := &Timeline{
		CacheDir: cacheDir,
		frames:   make([]*timeFrame, len(frames)),
		duration: duration,
		index:    -1,
	}

	for i, frame := range frames {
		t.frames[i] = &timeFrame{
			cache: New(maxBytes),
			frame: frame,
		}
	}

	if len(t.frames) > 0 {
		go t.init()
		// wait init step
		if since.Before(time.Now()) {
			time.Sleep(time.Microsecond)
		}
	}

	return t
}
