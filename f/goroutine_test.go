package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkPoolFuncJob(b *testing.B) {
	pool := f.NewPoolsFunc(func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret := pool.Process(10)
		if exp, act := 20, ret.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func BenchmarkPoolFuncTimedJob(b *testing.B) {
	pool := f.NewPoolsFunc(func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ret, err := pool.ProcessTimed(10, 2*time.Millisecond)
		if err != nil {
			b.Error(err)
		}
		if exp, act := 20, ret.(int); exp != act {
			b.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestPoolFuncJob(t *testing.T) {
	pool := f.NewPoolsFunc(func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for i := 0; i < 1000000; i++ {
		ret := pool.Process(10)
		if exp, act := 20, ret.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestPoolFuncJobTimed(t *testing.T) {
	pool := f.NewPoolsFunc(func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for i := 0; i < 1000000; i++ {
		ret, err := pool.ProcessTimed(10, 2*time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to process: %v", err)
		}
		if exp, act := 20, ret.(int); exp != act {
			t.Errorf("Wrong result: %v != %v", act, exp)
		}
	}
}

func TestPoolCallbackJob(t *testing.T) {
	pool := f.NewPoolsCallback()
	defer pool.Close()

	var counter int32
	for i := 0; i < 1000000; i++ {
		ret := pool.Process(func() {
			atomic.AddInt32(&counter, 1)
		})
		if ret != nil {
			t.Errorf("Non-nil callback response: %v", ret)
		}
	}

	ret := pool.Process("foo")
	if exp, act := f.ErrJobNotFunc, ret; exp != act {
		t.Errorf("Wrong result from non-func: %v != %v", act, exp)
	}

	if exp, act := int32(1000000), counter; exp != act {
		t.Errorf("Wrong result: %v != %v", act, exp)
	}
}

func TestPoolTimeout(t *testing.T) {
	f1 := f.NewFunc(func(in interface{}) interface{} {
		intVal := in.(int)
		<-time.After(time.Second)
		return intVal * 2
	})
	defer f1.Close()

	_, act := f1.ProcessTimed(1, time.Microsecond)
	if exp := f.ErrJobTimedOut; exp != act {
		t.Errorf("Wrong error returned: %v != %v", act, exp)
	}
}

func TestPoolTimedJobsAfterClose(t *testing.T) {
	f1 := f.NewFunc(func(in interface{}) interface{} {
		return in
	})
	f1.Close()

	_, act := f1.ProcessTimed(1, time.Duration(1))
	if exp := f.ErrPoolNotRunning; exp != act {
		t.Errorf("Wrong error returned: %v != %v", act, exp)
	}
}

func TestPoolJobsAfterClose(t *testing.T) {
	f1 := f.NewFunc(func(in interface{}) interface{} {
		return in
	})
	f1.Close()

	defer func() {
		if r := recover(); r != f.ErrPoolNotRunning {
			t.Errorf("Process after Stop() did not panic")
		}
	}()

	f1.Process(1)
}

func TestPoolParallelJobs(t *testing.T) {
	nWorkers := 100000

	testGroup := new(sync.WaitGroup)

	pool := f.NewPoolFunc(nWorkers, func(in interface{}) interface{} {
		intVal := in.(int)
		return intVal * 2
	})
	defer pool.Close()

	for j := 0; j < 1; j++ {
		testGroup.Add(nWorkers)

		for i := 0; i < nWorkers; i++ {
			go func() {
				ret := pool.Process(10)
				if exp, act := 20, ret.(int); exp != act {
					t.Errorf("Wrong result: %v != %v", act, exp)
				}
				testGroup.Done()
			}()
		}

		testGroup.Wait()
	}
}
