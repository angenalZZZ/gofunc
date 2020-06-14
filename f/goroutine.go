package f

import (
	"sync"
	"sync/atomic"
	"time"
)

// PoolWorker is spawning and managing a goroutine pool, allowing you
// to limit work coming from any number of goroutines with a synchronous API.
// an interface representing a working agent.
type PoolWorker interface {
	// Process will synchronously perform a job and return the result.
	Process(interface{}) interface{}

	// BlockUntilReady is called before each job is processed and must block the
	// calling goroutine until the Worker is ready to process the next job.
	BlockUntilReady()

	// Interrupt is called when a job is cancelled. The worker is responsible
	// for unblocking the Process implementation.
	Interrupt()

	// Terminate is called when a Worker is removed from the processing pool
	// and is responsible for cleaning up any held resources.
	Terminate()
}

// closurePoolWorker is a minimal Worker implementation that simply wraps a
// func(interface{}) interface{}
type closurePoolWorker struct {
	processor func(interface{}) interface{}
}

func (w *closurePoolWorker) Process(payload interface{}) interface{} {
	return w.processor(payload)
}

func (w *closurePoolWorker) BlockUntilReady() {}
func (w *closurePoolWorker) Interrupt()       {}
func (w *closurePoolWorker) Terminate()       {}

// callbackPoolWorker is a minimal Worker implementation that attempts to cast
// each job into func() and either calls it if successful or returns
// ErrJobNotFunc.
type callbackPoolWorker struct{}

func (w *callbackPoolWorker) Process(payload interface{}) interface{} {
	f, ok := payload.(func())
	if !ok {
		return ErrJobNotFunc
	}
	f()
	return nil
}

func (w *callbackPoolWorker) BlockUntilReady() {}
func (w *callbackPoolWorker) Interrupt()       {}
func (w *callbackPoolWorker) Terminate()       {}

// Pool is a struct that manages a collection of workers, each with their own
// goroutine. The Pool can initialize, expand, compress and close the workers,
// as well as processing jobs with the workers synchronously.
type Pool struct {
	queuedJobs int64

	ctor    func() PoolWorker
	workers []*poolWorkerWrapper
	reqChan chan poolWorkRequest

	workerMut sync.Mutex
}

// New creates a new Pool of workers that starts with n workers. You must
// provide a constructor function that creates new Worker types and when you
// change the size of the pool the constructor will be called to create each new
// Worker.
func NewPool(n int, ctor func() PoolWorker) *Pool {
	p := &Pool{
		ctor:    ctor,
		reqChan: make(chan poolWorkRequest),
	}
	p.SetSize(n)

	return p
}

// NewPoolFunc creates a new Pool of workers where each worker will process using
// the provided func.
func NewPoolFunc(n int, f func(interface{}) interface{}) *Pool {
	return NewPool(n, func() PoolWorker {
		return &closurePoolWorker{
			processor: f,
		}
	})
}

// NewPoolCallback creates a new Pool of workers where workers cast the job payload
// into a func() and runs it, or returns ErrNotFunc if the cast failed.
func NewPoolCallback(n int) *Pool {
	return NewPool(n, func() PoolWorker {
		return &callbackPoolWorker{}
	})
}

// Process will use the Pool to process a payload and synchronously return the
// result. Process can be called safely by any goroutines, but will panic if the
// Pool has been stopped.
func (p *Pool) Process(payload interface{}) interface{} {
	atomic.AddInt64(&p.queuedJobs, 1)

	request, open := <-p.reqChan
	if !open {
		panic(ErrPoolNotRunning)
	}

	request.jobChan <- payload

	payload, open = <-request.retChan
	if !open {
		panic(ErrWorkerClosed)
	}

	atomic.AddInt64(&p.queuedJobs, -1)
	return payload
}

// ProcessTimed will use the Pool to process a payload and synchronously return
// the result. If the timeout occurs before the job has finished the worker will
// be interrupted and ErrJobTimedOut will be returned. ProcessTimed can be
// called safely by any goroutines.
func (p *Pool) ProcessTimed(
	payload interface{},
	timeout time.Duration,
) (interface{}, error) {
	atomic.AddInt64(&p.queuedJobs, 1)
	defer atomic.AddInt64(&p.queuedJobs, -1)

	tout := time.NewTimer(timeout)

	var request poolWorkRequest
	var open bool

	select {
	case request, open = <-p.reqChan:
		if !open {
			return nil, ErrPoolNotRunning
		}
	case <-tout.C:
		return nil, ErrJobTimedOut
	}

	select {
	case request.jobChan <- payload:
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimedOut
	}

	select {
	case payload, open = <-request.retChan:
		if !open {
			return nil, ErrWorkerClosed
		}
	case <-tout.C:
		request.interruptFunc()
		return nil, ErrJobTimedOut
	}

	tout.Stop()
	return payload, nil
}

// QueueLength returns the current count of pending queued jobs.
func (p *Pool) QueueLength() int64 {
	return atomic.LoadInt64(&p.queuedJobs)
}

// SetSize changes the total number of workers in the Pool. This can be called
// by any goroutine at any time unless the Pool has been stopped, in which case
// a panic will occur.
func (p *Pool) SetSize(n int) {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	lWorkers := len(p.workers)
	if lWorkers == n {
		return
	}

	// Add extra workers if N > len(workers)
	for i := lWorkers; i < n; i++ {
		p.workers = append(p.workers, newPoolWorkerWrapper(p.reqChan, p.ctor()))
	}

	// Asynchronously stop all workers > N
	for i := n; i < lWorkers; i++ {
		p.workers[i].stop()
	}

	// Synchronously wait for all workers > N to stop
	for i := n; i < lWorkers; i++ {
		p.workers[i].join()
	}

	// Remove stopped workers from slice
	p.workers = p.workers[:n]
}

// GetSize returns the current size of the pool.
func (p *Pool) GetSize() int {
	p.workerMut.Lock()
	defer p.workerMut.Unlock()

	return len(p.workers)
}

// Close will terminate all workers and close the job channel of this Pool.
func (p *Pool) Close() {
	p.SetSize(0)
	close(p.reqChan)
}

// poolWorkRequest is a struct containing context representing a workers intention
// to receive a work payload.
type poolWorkRequest struct {
	// jobChan is used to send the payload to this worker.
	jobChan chan<- interface{}

	// retChan is used to read the result from this worker.
	retChan <-chan interface{}

	// interruptFunc can be called to cancel a running job. When called it is no
	// longer necessary to read from retChan.
	interruptFunc func()
}

// poolWorkerWrapper takes a Worker implementation and wraps it within a goroutine
// and channel arrangement. The poolWorkerWrapper is responsible for managing the
// lifetime of both the Worker and the goroutine.
type poolWorkerWrapper struct {
	worker        PoolWorker
	interruptChan chan struct{}

	// reqChan is NOT owned by this type, it is used to send requests for work.
	reqChan chan<- poolWorkRequest

	// closeChan can be closed in order to cleanly shutdown this worker.
	closeChan chan struct{}

	// closedChan is closed by the run() goroutine when it exits.
	closedChan chan struct{}
}

func newPoolWorkerWrapper(
	reqChan chan<- poolWorkRequest,
	worker PoolWorker,
) *poolWorkerWrapper {
	w := poolWorkerWrapper{
		worker:        worker,
		interruptChan: make(chan struct{}),
		reqChan:       reqChan,
		closeChan:     make(chan struct{}),
		closedChan:    make(chan struct{}),
	}

	go w.run()

	return &w
}

func (w *poolWorkerWrapper) interrupt() {
	close(w.interruptChan)
	w.worker.Interrupt()
}

func (w *poolWorkerWrapper) run() {
	jobChan, retChan := make(chan interface{}), make(chan interface{})
	defer func() {
		w.worker.Terminate()
		close(retChan)
		close(w.closedChan)
	}()

	for {
		// NOTE: Blocking here will prevent the worker from closing down.
		w.worker.BlockUntilReady()
		select {
		case w.reqChan <- poolWorkRequest{
			jobChan:       jobChan,
			retChan:       retChan,
			interruptFunc: w.interrupt,
		}:
			select {
			case payload := <-jobChan:
				result := w.worker.Process(payload)
				select {
				case retChan <- result:
				case <-w.interruptChan:
					w.interruptChan = make(chan struct{})
				}
			case _, _ = <-w.interruptChan:
				w.interruptChan = make(chan struct{})
			}
		case <-w.closeChan:
			return
		}
	}
}

func (w *poolWorkerWrapper) stop() {
	close(w.closeChan)
}

func (w *poolWorkerWrapper) join() {
	<-w.closedChan
}
