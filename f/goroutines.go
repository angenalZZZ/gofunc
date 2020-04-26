package f

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
)

// GoPool ants.Pool
type GoPool struct {
	*ants.Pool
}

// GoHandle a Simple Goroutines Pool
var GoHandle *GoPool

func init() {
	defaultPool, _ := ants.NewPool(ants.DefaultAntsPoolSize, ants.WithOptions(ants.Options{
		ExpiryDuration:   ants.DefaultCleanIntervalTime,
		PreAlloc:         false,
		Nonblocking:      true,
		MaxBlockingTasks: 0,
		PanicHandler: func(err interface{}) {
			_ = fmt.Errorf(" GoHandle/worker: %s\n %v", Now().LocalTimeString(), err)
		},
	}))
	GoHandle = &GoPool{Pool: defaultPool}
}

// GoWithFunc Create a ants.PoolWithFunc.
func GoWithFunc(size int, pf func(interface{}), options ...ants.Option) (*GoPoolWithFunc, error) {
	pool, err := ants.NewPoolWithFunc(size, pf, options...)
	if err != nil {
		return nil, err
	}
	return &GoPoolWithFunc{PoolWithFunc: pool}, nil
}

// GoTimes Create a ants.PoolWithFunc.
func GoTimes(n int, fn func(i int), options ...ants.Option) error {
	var size int
	if n < 100 {
		size = 10
	} else if n < 1000 {
		size = 100
	} else if n < 10000 {
		size = 1000
	} else {
		size = 2000
	}

	var wg sync.WaitGroup
	pool, err := ants.NewPoolWithFunc(size, func(i interface{}) {
		fn(i.(int))
		wg.Done()
	}, options...)
	if err != nil {
		return err
	}

	p := &GoPoolWithFunc{PoolWithFunc: pool}
	defer p.Release()
	for i := 0; i < n; i++ {
		wg.Add(1)
		_ = p.Invoke(i)
	}
	wg.Wait()
	return nil
}
