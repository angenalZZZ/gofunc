package f

import (
	"github.com/panjf2000/ants/v2"
	"time"
)

// GoPoolWithFunc ants.PoolWithFunc
type GoPoolWithFunc struct {
	*ants.PoolWithFunc
}

// GoHttpHandle a Http Handle Goroutines Pool
var GoHttpHandle *GoPoolWithFunc

func init() {
	defaultHttpHandlePool, _ := ants.NewPoolWithFunc(100000, func(payload interface{}) {
		if GoHttpHandle == nil {
			return
		}

		request, ok := payload.(*GoHttpHandleRequest)
		if !ok {
			return
		}

		request.HandleFunc()
	}, ants.WithOptions(ants.Options{
		ExpiryDuration:   time.Minute,
		PreAlloc:         true,
		MaxBlockingTasks: 0,
		Nonblocking:      true,
		PanicHandler: func(err interface{}) {
			panic(err)
		},
	}))
	GoHttpHandle = &GoPoolWithFunc{PoolWithFunc: defaultHttpHandlePool}
}

// GoHttpHandleRequest Handling input and output
type GoHttpHandleRequest struct {
	Body       []byte           // Input
	Param      interface{}      // Input
	HandleFunc func()           // HandleFunc: SetResult(result)
	_result    chan interface{} // Output
}

// SetResult Write to a output channel.
func (g *GoHttpHandleRequest) SetResult(result interface{}) {
	if g._result == nil {
		g._result = make(chan interface{}, 1)
	}
	g._result <- result
}

// GetResult Read from a output channel.
func (g *GoHttpHandleRequest) GetResult() interface{} {
	if g._result == nil {
		return nil
	}
	return <-g._result
}

// Invoke Handling logic, return throttle limit error.
// Throttle the requests traffic with ants pool. This process is asynchronous and
// you can receive a result from the channel defined outside.
func (g *GoHttpHandleRequest) Invoke() error {
	return GoHttpHandle.Invoke(g)
}

// NewHttpHandleRequest Create a http request handler.
func NewHttpHandleRequest(param interface{}) *GoHttpHandleRequest {
	return &GoHttpHandleRequest{
		Body:       nil,
		Param:      param,
		HandleFunc: nil,
		_result:    make(chan interface{}, 1),
	}
}

// NewHttpHandleRequestBody Create a http request handler.
func NewHttpHandleRequestBody(body []byte) *GoHttpHandleRequest {
	return &GoHttpHandleRequest{
		Body:       body,
		Param:      nil,
		HandleFunc: nil,
		_result:    make(chan interface{}, 1),
	}
}

// GoWithFunc Create a ants.PoolWithFunc.
func GoWithFunc(size int, pf func(interface{}), options ...ants.Option) (*GoPoolWithFunc, error) {
	pool, err := ants.NewPoolWithFunc(size, pf, options...)
	if err != nil {
		return nil, err
	}
	return &GoPoolWithFunc{PoolWithFunc: pool}, nil
}
