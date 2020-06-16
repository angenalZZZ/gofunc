package f_test

import (
	"github.com/angenalZZZ/gofunc/data/random"
	"github.com/angenalZZZ/gofunc/f"
	"testing"
)

func TestCMap_JSON(t *testing.T) {
	m := f.NewConcurrentMap()
	m.MSet(map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	})
	s, _ := m.JSON()
	t.Logf("%s\n", s)

	m2, err2 := f.NewConcurrentMapFromJSON(s)
	if err2 != nil {
		t.Fatal(err2)
	}
	s2, _ := m2.JSON()
	t.Logf("%s\n", s2)
}

// go test -v -cpu=4 -benchtime=15s -benchmem -bench=^BenchmarkCMap_Set$ -test.run ^none$ ./f
// go test -c -o %TEMP%\t01.exe ./f && %TEMP%\t01.exe -test.v -test.bench ^BenchmarkCMap_Set$ -test.run ^none$
func BenchmarkCMap_Set(b *testing.B) {
	b.StopTimer()
	m := f.NewConcurrentMap()
	k := random.AlphaNumber(32)
	v := random.AlphaNumber(1024) // every time 1kB data request: cpu=4 5500k/qps 0.1ms/op
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		m.Set(k, v)
	}
}
