package snowid

import (
	"math"
	"testing"
	"time"
)

func TestDefaultIdGenerator(t *testing.T) {
	i1, i2 := NextId(), NextId()
	if i1 == i2 {
		t.Fail()
	}
}

func TestNewIdGenerator(t *testing.T) {
	var options = NewIdGeneratorOptions(1)
	options.BaseTime = time.Now().Add(time.Hour*-1).UnixNano() / 1e6
	options.WorkerIdBitLength = 9
	options.SeqBitLength = 9
	var times = 1000000 // 与 WorkerIdBitLength 有关系

	SetDefaultIdGenerator(options)

	// start benchmark test
	t1 := time.Now()

	for i := 0; i < times; i++ {
		NextId()
	}

	t2 := time.Now()
	ts := t2.Sub(t1)
	qps := times * 1000 / int(math.Max(float64(1), float64(ts.Milliseconds())))
	t.Logf("Take time %s and %d qps, total %d times", ts, qps, times)
}
