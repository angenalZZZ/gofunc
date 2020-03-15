package f

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoroutines(t *testing.T) {
	defer GoHandle.Release()

	var sum int32
	myFunc := func(i interface{}) {
		n := i.(int32)
		atomic.AddInt32(&sum, n)
	}
	demoFunc := func() {
		time.Sleep(10 * time.Millisecond)
	}

	// Use the common pool.
	runTimes := 1000
	var wg sync.WaitGroup
	syncCalculateSum := func() {
		demoFunc()
		wg.Done()
	}
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		_ = GoHandle.Submit(syncCalculateSum)
	}
	wg.Wait()
	goroutines := GoHandle.Running()
	t.Logf("running goroutines: %d\n", goroutines)
	t.Log("finish all tasks.")

	// Use the pool with a function,
	// set 10 to the capacity of goroutine pool and 1 second for expired duration.
	p, _ := GoWithFunc(20, func(i interface{}) {
		myFunc(i)
		wg.Done()
	})
	defer p.Release()
	// Submit tasks one by one.
	for i := 0; i < runTimes; i++ {
		wg.Add(1)
		_ = p.Invoke(int32(i))
	}
	wg.Wait()
	t.Logf("running goroutines: %d\n", p.Running())
	t.Logf("finish all tasks, result is %d\n", sum)
}
