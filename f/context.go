package f

import (
	"context"
	"os"
	"os/signal"
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
