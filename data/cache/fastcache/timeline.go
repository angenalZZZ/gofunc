package fastcache

import (
	"fmt"
	"github.com/angenalZZZ/gofunc/f"
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
	cache *Cache
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

func (c *timeFrame) filename() string {
	return fmt.Sprintf("%s.%d", c.frame.Since.LocalTimeStampString(true), c.index)
}

func (c *timeFrame) save(cacheDir string) {
	time.Sleep(time.Second)
	if c.index == 0 {
		return
	}
	filePath := filepath.Join(cacheDir, c.filename())
	_ = c.cache.SaveToFileConcurrent(filePath, 0)
}

func (t *Timeline) init() {
	l := len(t.frames)
	if l == 0 {
		return
	}

	p := int64(t.duration.Seconds())
	m, n := t.frames[l-1].frame.Until.UnixSecond, t.frames[0].frame.Since.UnixSecond
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

	go t.init()
	return t
}
