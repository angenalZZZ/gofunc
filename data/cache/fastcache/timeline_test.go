package fastcache

import (
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"time"
)

const testTimelineCacheDir string = `A:\test` // or set f.CurrentDir()

func TestTimelineInit(t *testing.T) {
	tl := NewTimeline(time.Now(), time.Now().Add(50*time.Second), 5*time.Second, testTimelineCacheDir, 1024)
	t.Logf("%3d: %s", tl.index, f.NowLocalString(true))
	for index := int64(0); tl.index != -1; {
		for index == tl.index {
			time.Sleep(time.Microsecond)
		}
		if tl.index != -1 {
			t.Logf("%3d: %s", tl.index, f.NowLocalString(true))
			index = tl.index
		} else {
			t.Logf("end: %s", f.NowLocalString(true))
		}
	}
}

func TestTimelineReader(t *testing.T) {
	oldFiles, _ := filepath.Glob(filepath.Join(testTimelineCacheDir, "*"))
	sort.Strings(oldFiles)
	for _, oldFile := range oldFiles {
		_, f := filepath.Split(oldFile)
		if ok, _ := regexp.MatchString(`^\d{10,}\.\d+`, f); !ok {
			continue
		}
		t.Log(oldFile)
		//s := strings.Split(f, ".")
		//start, _ := strconv.ParseInt(s[0], 10, 0)
		//index, _ := strconv.ParseInt(s[1], 10, 0)
		//cache, err := fastcache.LoadFromFile(oldFile)
		//if err != nil || cache == nil {
		//	continue
		//}
		//writer := &CacheWriter{
		//	Cache: cache,
		//	Start: start,
		//	Index: uint32(index),
		//}
	}
}

// go test -v -cpu=4 -benchtime=10s -benchmem -bench=^BenchmarkTimelineWriter -run ^none$ github.com/angenalZZZ/gofunc/data/cache/fastcache
// go test -c -o %TEMP%\t01.exe github.com/angenalZZZ/gofunc/data/cache/fastcache && %TEMP%\t01.exe -test.v -test.bench ^BenchmarkTimelineWriter -test.run ^none$
func BenchmarkTimelineWriter(b *testing.B) {
	b.StopTimer()
	tl := NewTimeline(time.Now(), time.Now().Add(50*time.Second), 5*time.Second, testTimelineCacheDir, 1024)
	time.Sleep(time.Microsecond)
	//l := 5120 // every time 5kB data request: cpu=4 1200k/qps 0.8ms/op
	//l := 2018 // every time 2kB data request: cpu=4 1800k/qps 0.5ms/op
	//l := 1024 // every time 1kB data request: cpu=4 2400k/qps 0.4ms/op
	l := 128 // every time 128B data request: cpu=4 3200k/qps 0.3ms/op
	p := []byte(random.AlphaNumberLower(l))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tl.Write(p)
	}
}
