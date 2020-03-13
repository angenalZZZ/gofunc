package f

import "github.com/panjf2000/ants/v2"

type GoPool struct {
	*ants.Pool
}
type GoPoolWithFunc struct {
	*ants.PoolWithFunc
}

var (
	Go     *GoPool
	GoHttp *GoPoolWithFunc
)

type GoHttpRequest struct {
	Body       []byte
	Param      interface{}
	HandleFunc func() // TODO: SetResult(result)
	result     chan interface{}
}

// NewHttpRequest New HttpRequest.
func NewHttpRequest(param interface{}, body []byte) *GoHttpRequest {
	return &GoHttpRequest{
		Body:       body,
		Param:      param,
		HandleFunc: nil,
		result:     make(chan interface{}, 1),
	}
}

// DoHttpRequest Do something in a goroutine pool.
func DoHttpRequest(request *GoHttpRequest) interface{} {
	if request == nil || request.HandleFunc == nil {
		return nil
	}
	if err := GoHttp.Invoke(request); err != nil {
		return err
	}
	return request.GetResult()
}

// NewPoolWithFunc New GoPoolWithFunc.
func NewPoolWithFunc(size int, pf func(interface{}), options ...ants.Option) (*GoPoolWithFunc, error) {
	pool, err := ants.NewPoolWithFunc(size, pf, options...)
	if err != nil {
		return nil, err
	}
	return &GoPoolWithFunc{PoolWithFunc: pool}, nil
}

func (g *GoHttpRequest) SetResult(result interface{}) {
	if g.result == nil {
		g.result = make(chan interface{}, 1)
	}
	g.result <- result
}

func (g *GoHttpRequest) GetResult() interface{} {
	if g.result == nil {
		return nil
	}
	return <-g.result
}

func init() {
	defaultPool, _ := ants.NewPool(ants.DefaultAntsPoolSize)
	Go = &GoPool{Pool: defaultPool}

	defaultHttpHandlePool, _ := ants.NewPoolWithFunc(100000, func(payload interface{}) {
		if GoHttp == nil {
			return
		}

		request, ok := payload.(*GoHttpRequest)
		if !ok {
			return
		}

		request.HandleFunc()
	})
	GoHttp = &GoPoolWithFunc{PoolWithFunc: defaultHttpHandlePool}
}
