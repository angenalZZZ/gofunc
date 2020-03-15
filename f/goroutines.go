package f

import (
	"github.com/panjf2000/ants/v2"
)

// GoPool ants.Pool
type GoPool struct {
	*ants.Pool
}

// GoHandle a Simple Goroutines Pool
var GoHandle *GoPool

func init() {
	defaultPool, _ := ants.NewPool(ants.DefaultAntsPoolSize)
	GoHandle = &GoPool{Pool: defaultPool}
}
