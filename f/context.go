package f

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// ContextOnInterrupt creates a new context that cancels on SIGINT or SIGTERM.
func ContextOnInterrupt() (context.Context, func()) {
	return ContextWrapSignal(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

// ContextOnSignal creates a new context that cancels on the given signals.
func ContextOnSignal(signals ...os.Signal) (context.Context, func()) {
	return ContextWrapSignal(context.Background(), signals...)
}

// ContextWrapSignal creates a new context that cancels on the given signals. It wraps the provided context.
func ContextWrapSignal(parent context.Context, signals ...os.Signal) (ctx context.Context, closer func()) {
	ctx, closer = context.WithCancel(parent)

	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)

	go func() {
		select {
		case <-c:
			closer()
		case <-ctx.Done():
		}
	}()

	return ctx, closer
}

// contextWaitChan is an unexported type for channel used as the blocking statement in
// the waitFunc returned by wait.WithWait.
type contextWaitChan chan struct{}

// contextWaitKey is the key for wait.contextWaitChan values in Contexts. It is unexported;
// clients use wait.WithWait and wait.Done instead of using this key directly.
const contextWaitKey = 32 << (uint64(^uint(0)) >> 63)

// contextWaitMutex is the mutual exclusion lock used to prevent a race condition in wait.Done.
var contextWaitMutex sync.Mutex

// ContextWithWait creates a new context that wait on the given ContextDone().
func ContextWithWait(parent context.Context) (ctx context.Context, waitFunc func()) {
	ctx = parent

	var ok bool
	var wait contextWaitChan
	if wait, ok = ctx.Value(contextWaitKey).(contextWaitChan); !ok {
		wait = make(contextWaitChan)
		ctx = context.WithValue(parent, contextWaitKey, wait)
	}

	waitFunc = func() {
		select {
		case <-wait:
		}
	}
	return
}

// DoneContext unblocks the waitFunc returned by wait.WithWait for the provided ctx, if available.
// If waitFunc has already been unblocked, then nothing happens.
func DoneContext(ctx context.Context) {
	if wait, ok := ctx.Value(contextWaitKey).(contextWaitChan); ok {
		contextWaitMutex.Lock()
		defer contextWaitMutex.Unlock()

		select {
		case <-wait:
		default:
			close(wait)
		}
	}
}
