package f

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
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
