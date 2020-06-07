package queue_test

import (
	"github.com/angenalZZZ/gofunc/data/random"
	"io/ioutil"
	"os"
	"testing"

	"github.com/angenalZZZ/gofunc/data/queue"
)

// go test -v -cpu=4 -benchtime=15s -benchmem -bench=^BenchmarkQueueEnqueue$ -run ^none$ github.com/angenalZZZ/gofunc/data/queue
func BenchmarkQueueEnqueue(b *testing.B) {
	b.StopTimer()
	// Open/create a queue.
	dataDir, _ := ioutil.TempDir("", "")
	q, err := queue.OpenQueue(dataDir)
	if err != nil {
		b.Error(err)
		return
	}
	defer func() {
		_ = q.Drop()
		_ = q.Close()
		_ = os.RemoveAll(dataDir)
	}()
	//l := 1024 // every time 1kB data request: cpu=4 70k/qps 8ms/op
	l := 128 // every time 128B data request: cpu=4 75k/qps 8ms/op
	p := []byte(random.AlphaNumberLower(l))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Enqueue an item.
		_, err := q.Enqueue(p)
		if err != nil {
			b.Error(err)
			return
		}
	}
}
