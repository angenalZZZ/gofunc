package f_test

import (
	"context"
	"github.com/angenalZZZ/gofunc/f"
	"log"
	"net/http"
	"time"
)

//func TestContextWrapSignal(t *testing.T) {
//	ctx, cancel := f.ContextWrapSignal(context.Background(), syscall.SIGUSR1)
//	defer cancel()
//
//	select {
//	case <-ctx.Done():
//		t.Fatal("context should not be done")
//	case <-time.After(10 * time.Millisecond):
//	}
//
//	if err := syscall.Kill(syscall.Getpid(), syscall.SIGUSR1); err != nil {
//		t.Fatal("failed to signal")
//	}
//
//	select {
//	case <-ctx.Done():
//		// expected
//	case <-time.After(10 * time.Millisecond):
//		t.Fatal("context should have been done")
//	}
//}

func ExampleContextOnInterrupt() {
	ctx, cancel := f.ContextOnInterrupt()
	defer cancel()

	s := &http.Server{
		Addr: ":8080",
	}
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for CTRL+C
	<-ctx.Done()

	// Stop the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
