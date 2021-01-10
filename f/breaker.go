package f

// Implements the Circuit Breaker pattern.
// See https://msdn.microsoft.com/en-us/library/dn589784.aspx.

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// BreakerState is a type that represents a state of CircuitBreaker.
type BreakerState int

// These constants are states of CircuitBreaker.
const (
	BreakerStateClosed BreakerState = iota
	BreakerStateHalfOpen
	BreakerStateOpen
)

var (
	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests
	ErrTooManyRequests = errors.New("too many requests")
	// ErrBreakerOpenState is returned when the CB state is open
	ErrBreakerOpenState = errors.New("circuit breaker is open")
)

// String implements stringer interface.
func (s BreakerState) String() string {
	switch s {
	case BreakerStateClosed:
		return "closed"
	case BreakerStateHalfOpen:
		return "half-open"
	case BreakerStateOpen:
		return "open"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}

// BreakerCounts holds the numbers of requests and their successes/failures.
// CircuitBreaker clears the internal BreakerCounts either
// on the change of the state or at the closed-state intervals.
// BreakerCounts ignores the results of the requests sent before clearing.
type BreakerCounts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

func (c *BreakerCounts) onRequest() {
	c.Requests++
}

func (c *BreakerCounts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *BreakerCounts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *BreakerCounts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// BreakerSettings configures CircuitBreaker:
//
// Name is the name of the CircuitBreaker.
//
// MaxRequests is the maximum number of requests allowed to pass through
// when the CircuitBreaker is half-open.
// If MaxRequests is 0, the CircuitBreaker allows only 1 request.
//
// Interval is the cyclic period of the closed state
// for the CircuitBreaker to clear the internal BreakerCounts.
// If Interval is less than or equal to 0, the CircuitBreaker doesn't clear internal BreakerCounts during the closed state.
//
// Timeout is the period of the open state,
// after which the state of the CircuitBreaker becomes half-open.
// If Timeout is less than or equal to 0, the timeout value of the CircuitBreaker is set to 60 seconds.
//
// ReadyToTrip is called with a copy of BreakerCounts whenever a request fails in the closed state.
// If ReadyToTrip returns true, the CircuitBreaker will be placed into the open state.
// If ReadyToTrip is nil, default ReadyToTrip is used.
// Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
//
// OnStateChange is called whenever the state of the CircuitBreaker changes.
//
// IsSuccessful is called with the error returned from the request, if not nil.
// If IsSuccessful returns false, the error is considered a failure, and is counted towards tripping the circuit breaker.
// If IsSuccessful returns true, the error will be returned to the caller without tripping the circuit breaker.
// If IsSuccessful is nil, default IsSuccessful is used, which returns false for all non-nil errors.
type BreakerSettings struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts BreakerCounts) bool
	OnStateChange func(name string, from BreakerState, to BreakerState)
	IsSuccessful  func(err error) bool
}

// CircuitBreaker is a state machine to prevent sending requests that are likely to fail.
type CircuitBreaker struct {
	name          string
	maxRequests   uint32
	interval      time.Duration
	timeout       time.Duration
	readyToTrip   func(counts BreakerCounts) bool
	isSuccessful  func(err error) bool
	onStateChange func(name string, from BreakerState, to BreakerState)

	mutex      sync.Mutex
	state      BreakerState
	generation uint64
	counts     BreakerCounts
	expiry     time.Time
}

// TwoStepCircuitBreaker is like CircuitBreaker but instead of surrounding a function
// with the breaker functionality, it only checks whether a request can proceed and
// expects the caller to report the outcome in a separate step using a callback.
type TwoStepCircuitBreaker struct {
	cb *CircuitBreaker
}

// NewCircuitBreaker returns a new CircuitBreaker configured with the given BreakerSettings.
func NewCircuitBreaker(st BreakerSettings) *CircuitBreaker {
	cb := new(CircuitBreaker)

	cb.name = st.Name
	cb.onStateChange = st.OnStateChange

	if st.MaxRequests == 0 {
		cb.maxRequests = 1
	} else {
		cb.maxRequests = st.MaxRequests
	}

	if st.Interval <= 0 {
		cb.interval = defaultBreakerInterval
	} else {
		cb.interval = st.Interval
	}

	if st.Timeout <= 0 {
		cb.timeout = defaultBreakerTimeout
	} else {
		cb.timeout = st.Timeout
	}

	if st.ReadyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	} else {
		cb.readyToTrip = st.ReadyToTrip
	}

	if st.IsSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	} else {
		cb.isSuccessful = st.IsSuccessful
	}

	cb.toNewGeneration(time.Now())

	return cb
}

// NewTwoStepCircuitBreaker returns a new TwoStepCircuitBreaker configured with the given BreakerSettings.
func NewTwoStepCircuitBreaker(st BreakerSettings) *TwoStepCircuitBreaker {
	return &TwoStepCircuitBreaker{
		cb: NewCircuitBreaker(st),
	}
}

const defaultBreakerInterval = time.Duration(0) * time.Second
const defaultBreakerTimeout = time.Duration(60) * time.Second

func defaultReadyToTrip(counts BreakerCounts) bool {
	return counts.ConsecutiveFailures > 5
}

func defaultIsSuccessful(err error) bool {
	return err == nil
}

// Name returns the name of the CircuitBreaker.
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State returns the current state of the CircuitBreaker.
func (cb *CircuitBreaker) State() BreakerState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// Execute runs the given request if the CircuitBreaker accepts it.
// Execute returns an error instantly if the CircuitBreaker rejects the request.
// Otherwise, Execute returns the result of the request.
// If a panic occurs in the request, the CircuitBreaker handles it as an error
// and causes the same panic again.
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	result, err := req()
	cb.afterRequest(generation, cb.isSuccessful(err))
	return result, err
}

// Name returns the name of the TwoStepCircuitBreaker.
func (cb *TwoStepCircuitBreaker) Name() string {
	return cb.cb.Name()
}

// State returns the current state of the TwoStepCircuitBreaker.
func (cb *TwoStepCircuitBreaker) State() BreakerState {
	return cb.cb.State()
}

// Allow checks if a new request can proceed. It returns a callback that should be used to
// register the success or failure in a separate step. If the circuit breaker doesn't allow
// requests, it returns an error.
func (cb *TwoStepCircuitBreaker) Allow() (done func(success bool), err error) {
	generation, err1 := cb.cb.beforeRequest()
	if err1 != nil {
		return nil, err1
	}

	return func(success bool) {
		cb.cb.afterRequest(generation, success)
	}, nil
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == BreakerStateOpen {
		return generation, ErrBreakerOpenState
	} else if state == BreakerStateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrTooManyRequests
	}

	cb.counts.onRequest()
	return generation, nil
}

func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state BreakerState, now time.Time) {
	switch state {
	case BreakerStateClosed:
		cb.counts.onSuccess()
	case BreakerStateHalfOpen:
		cb.counts.onSuccess()
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(BreakerStateClosed, now)
		}
	}
}

func (cb *CircuitBreaker) onFailure(state BreakerState, now time.Time) {
	switch state {
	case BreakerStateClosed:
		cb.counts.onFailure()
		if cb.readyToTrip(cb.counts) {
			cb.setState(BreakerStateOpen, now)
		}
	case BreakerStateHalfOpen:
		cb.setState(BreakerStateOpen, now)
	}
}

func (cb *CircuitBreaker) currentState(now time.Time) (BreakerState, uint64) {
	switch cb.state {
	case BreakerStateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case BreakerStateOpen:
		if cb.expiry.Before(now) {
			cb.setState(BreakerStateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

func (cb *CircuitBreaker) setState(state BreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case BreakerStateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case BreakerStateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // BreakerStateHalfOpen
		cb.expiry = zero
	}
}
