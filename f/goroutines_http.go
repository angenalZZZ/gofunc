package f

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
)

// GoPoolWithFunc ants.PoolWithFunc.
type GoPoolWithFunc struct {
	*ants.PoolWithFunc
}

// GoHttpHandle a Http Handle Goroutines Pool.
var GoHttpHandle *GoPoolWithFunc

func init() {
	GoHttpHandle = NewHttpHandlePool(100000)
}

// GoHttpHandleRequest Handling input and output.
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
func (g *GoHttpHandleRequest) Invoke(pool *GoPoolWithFunc) error {
	return pool.Invoke(g)
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

// NewHttpHandlePool Create a Http Handle Goroutines Pool.
// @poolTotalSize 100000: SetHeader total goroutines and tasks.
func NewHttpHandlePool(poolTotalSize int, options ...ants.Options) *GoPoolWithFunc {
	var option ants.Options
	if len(options) > 0 {
		option = options[0]
	} else {
		option = ants.Options{
			ExpiryDuration:   ants.DefaultCleanIntervalTime,
			PreAlloc:         true,
			Nonblocking:      true,
			MaxBlockingTasks: 0,
			PanicHandler: func(err interface{}) {
				_ = fmt.Errorf(" GoHttpHandle/worker: %s\n %v", Now().LocalTimeString(), err)
			},
		}
	}

	poolWithFunc, _ := ants.NewPoolWithFunc(poolTotalSize, func(payload interface{}) {
		request, ok := payload.(*GoHttpHandleRequest)
		if !ok {
			return
		}

		request.HandleFunc()
	}, ants.WithOptions(option))
	return &GoPoolWithFunc{PoolWithFunc: poolWithFunc}
}

// NewHttpHandlePoolWithThrottle Create a Http Handle Goroutines Pool.
// @throttleLimitNumber 1000: More than it, return HTTP code 429 Too Many Requests.
// @poolTotalSize 100000: SetHeader total goroutines and tasks.
func NewHttpHandlePoolWithThrottle(throttleLimitNumber, poolTotalSize int, options ...ants.Options) *GoPoolWithFunc {
	var option ants.Options
	if len(options) > 0 {
		option = options[0]
	} else {
		option = ants.Options{
			ExpiryDuration:   ants.DefaultCleanIntervalTime,
			PreAlloc:         true,
			Nonblocking:      false,
			MaxBlockingTasks: throttleLimitNumber,
			PanicHandler: func(err interface{}) {
				_ = fmt.Errorf(" GoHttpHandle/worker: %s\n %v", Now().LocalTimeString(), err)
			},
		}
	}

	poolWithFunc, _ := ants.NewPoolWithFunc(poolTotalSize, func(payload interface{}) {
		request, ok := payload.(*GoHttpHandleRequest)
		if !ok {
			return
		}

		request.HandleFunc()
	}, ants.WithOptions(option))
	return &GoPoolWithFunc{PoolWithFunc: poolWithFunc}
}
