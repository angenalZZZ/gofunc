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
	Frames   []*TimeFrame
	Duration time.Duration
	Index    int64
}

// TimeFrame time bounds on which data to retrieve.
type TimeFrame struct {
	Cache *Cache // a fast thread-safe inmemory cache optimized for big number of entries.
	Frame *f.TimeFrame
	Index uint32
}

func (t *Timeline) Write(p []byte) (n int, err error) {
	if t.Index == -1 {
		return
	}

	c := t.Frames[t.Index]
	i := atomic.AddUint32(&c.Index, 1)
	c.Cache.Set(f.BytesUint32(i), p)
	return int(i), nil
}

func (t *Timeline) Save() {
	for _, frame := range t.Frames {
		frame.Save(t.CacheDir)
	}
}

func (t *Timeline) Remove(index int) {
	if index >= 0 && index < len(t.Frames) {
		t.Frames[index].Remove(t.CacheDir)
	}
}

func (t *Timeline) RemoveAll() {
	for _, frame := range t.Frames {
		frame.Remove(t.CacheDir)
	}
}

func (c *TimeFrame) Dirname() string {
	return fmt.Sprintf("%s.%d", c.Frame.Since.LocalTimeStampString(true), c.Index)
}

func (c *TimeFrame) Filename() string {
	return fmt.Sprintf("%s.%d.json", c.Frame.Since.LocalTimeStampString(true), c.Index)
}

func (c *TimeFrame) Save(cacheDir string) {
	time.Sleep(time.Microsecond)
	if c.Index == 0 {
		return
	}

	fileStat := new(Stats)
	c.Cache.UpdateStats(fileStat)
	data, err := f.EncodeJson(fileStat)
	logErr := log.New(os.Stderr, "", 0)
	if err != nil {
		logErr.Print(err)
	}

	filePath := filepath.Join(cacheDir, c.Filename())
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		logErr.Print(err)
	}

	dirPath := filepath.Join(cacheDir, c.Dirname())
	if err = c.Cache.SaveToFileConcurrent(dirPath, 0); err != nil {
		logErr.Print(err)
	} else {
		c.Cache.Reset() // Reset removes all the items from the cache.
	}
}

func (c *TimeFrame) Remove(cacheDir string) {
	if c.Index == 0 {
		return
	}

	filePath := filepath.Join(cacheDir, c.Dirname())
	err := os.Remove(filePath + ".json")
	logErr := log.New(os.Stderr, "", 0)
	if err != nil {
		logErr.Print(err)
	}

	err = os.RemoveAll(filePath)
	if err != nil {
		logErr.Print(err)
	}
}

func (t *Timeline) init() {
	p := int64(t.Duration.Seconds())
	n := t.Frames[0].Frame.Since.UnixSecond
	m := t.Frames[len(t.Frames)-1].Frame.Until.UnixSecond
	for u := time.Now().Unix(); u < m; u++ {
		index := (u - n) / p
		if index >= 0 && index != t.Index {
			if t.Index != -1 {
				go t.Frames[t.Index].Save(t.CacheDir)
			}
			atomic.StoreInt64(&t.Index, index)
		}
		time.Sleep(time.Second)
	}
	t.Index = -1
}

func NewTimeline(since, until time.Time, duration time.Duration, cacheDir string, maxBytes int) *Timeline {
	frames := f.NewTimeFrames(since, until, duration)

	t := &Timeline{
		CacheDir: cacheDir,
		Frames:   make([]*TimeFrame, len(frames)),
		Duration: duration,
		Index:    -1,
	}

	for i, frame := range frames {
		t.Frames[i] = &TimeFrame{
			Cache: New(maxBytes),
			Frame: frame,
		}
	}

	if len(t.Frames) > 0 {
		go t.init()
		// wait init step
		if since.Before(time.Now()) {
			time.Sleep(time.Microsecond)
		}
	}

	return t
}
