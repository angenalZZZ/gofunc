package f

import "github.com/panjf2000/ants/v2"

type Gos struct {
	*ants.Pool
}

var Go *Gos

func init() {
	defaultPool, _ := ants.NewPool(ants.DefaultAntsPoolSize)
	Go = &Gos{Pool: defaultPool}
}

func GoWithFunc(size int, pf func(interface{}), options ...ants.Option) (*ants.PoolWithFunc, error) {
	return ants.NewPoolWithFunc(size, pf, options...)
}
