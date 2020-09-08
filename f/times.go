package f

import (
	"fmt"
	"io"
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

// Ops executed a number of operations over a multiple goroutines.
// count is the number of operations.
// threads is the number goroutines.
// op is the operation function
func Ops(count, threads int, op func(i, thread int)) {
	var start time.Time
	var wg sync.WaitGroup
	wg.Add(threads)
	output := Output
	if output != nil {
		start = time.Now()
	}
	for i := 0; i < threads; i++ {
		s, e := count/threads*i, count/threads*(i+1)
		if i == threads-1 {
			e = count
		}
		go func(i, s, e int) {
			defer wg.Done()
			for j := s; j < e; j++ {
				op(j, i)
			}
		}(i, s, e)
	}
	wg.Wait()

	if output != nil {
		dur := time.Since(start)
		var alloc uint64
		WriteOutput(output, count, threads, dur, alloc)
	}
}

func opsComma(n int) string {
	s1, s2 := fmt.Sprintf("%d", n), ""
	for i, j := len(s1)-1, 0; i >= 0; i, j = i-1, j+1 {
		if j%3 == 0 && j != 0 {
			s2 = "," + s2
		}
		s2 = string(s1[i]) + s2
	}
	return s2
}

func opsMem(alloc uint64) string {
	switch {
	case alloc <= 1024:
		return fmt.Sprintf("%d bytes", alloc)
	case alloc <= 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(alloc)/1024)
	case alloc <= 1024*1024*1024:
		return fmt.Sprintf("%.1f MB", float64(alloc)/1024/1024)
	default:
		return fmt.Sprintf("%.1f GB", float64(alloc)/1024/1024/1024)
	}
}

// WriteOutput writes an output line to the specified writer
func WriteOutput(w io.Writer, count, threads int, elapsed time.Duration, alloc uint64) {
	var ss string
	if threads != 1 {
		ss = fmt.Sprintf("over %d threads ", threads)
	}
	var qps int
	if count > 0 {
		qps = int(elapsed / time.Duration(count))
	}
	var allocStr string
	if alloc > 0 {
		var bops uint64
		if count > 0 {
			bops = alloc / uint64(count)
		}
		allocStr = fmt.Sprintf(", %s, %d bytes/op", opsMem(alloc), bops)
	}
	_, _ = fmt.Fprintf(w, "%s ops %sin %.0fms, %s/sec, %d ns/op%s\n",
		opsComma(count), ss, elapsed.Seconds()*1000,
		opsComma(int(float64(count)/elapsed.Seconds())),
		qps, allocStr,
	)
}
