package f

import (
	"sync"
	"time"
)

// At Runs fn at the specified time.
func At(t time.Time, fn func()) (done <-chan bool) {
	return After(t.Sub(time.Now()), fn)
}

// Until Runs until time in every dur.
func Until(t time.Time, dur time.Duration, fn func()) (done <-chan bool) {
	ch := make(chan bool, 1)
	untilRecv(ch, t, dur, fn)
	return ch
}

func untilRecv(ch chan bool, t time.Time, dur time.Duration, fn func()) {
	if t.Sub(time.Now()) > 0 {
		time.AfterFunc(dur, func() {
			fn()
			untilRecv(ch, t, dur, fn)
		})
		return
	}
	doneSig(ch, true)
}

// After Runs fn after duration. Similar to time.AfterFunc
func After(duration time.Duration, fn func()) (done <-chan bool) {
	ch := make(chan bool, 1)
	time.AfterFunc(duration, func() {
		fn()
		doneSig(ch, true)
	})
	return ch
}

// Every Runs fn in every specified duration.
func Every(dur time.Duration, fn func()) {
	time.AfterFunc(dur, func() {
		fn()
		Every(dur, fn)
	})
}

// Timeout Runs fn and times out if it runs longer than the provided
// duration. It will send false to the returning
// channel if timeout occurs.
func Timeout(duration time.Duration, fn func()) (done <-chan bool) {
	ch := make(chan bool, 2)
	go func() {
		<-time.After(duration)
		doneSig(ch, false)
	}()
	go func() {
		fn()
		doneSig(ch, true)
	}()
	return ch
}

// All Starts to run the given list of fns concurrently.
func All(fns ...func()) (done <-chan bool) {
	var wg sync.WaitGroup
	wg.Add(len(fns))

	ch := make(chan bool, 1)
	for _, fn := range fns {
		go func(f func()) {
			f()
			wg.Done()
		}(fn)
	}
	go func() {
		wg.Wait()
		doneSig(ch, true)
	}()
	return ch
}

// AllWithThrottle Starts to run the given list of fns concurrently,
// at most n fns at a time.
func AllWithThrottle(throttle int, fns ...func()) (done <-chan bool) {
	ch := make(chan bool, 1)
	go func() {
		for {
			num := throttle
			if throttle > len(fns) {
				num = len(fns)
			}
			next := fns[:num]
			fns = fns[num:]
			<-All(next...)
			if len(fns) == 0 {
				doneSig(ch, true)
				break
			}
		}
	}()
	return ch
}

// Replicate Run the same function with n copies.
func Replicate(n int, fn func()) (done <-chan bool) {
	fns := make([]func(), n)
	for i := 0; i < n; i++ {
		fns[i] = fn
	}
	return All(fns...)
}

func doneSig(ch chan bool, val bool) {
	ch <- val
	close(ch)
}