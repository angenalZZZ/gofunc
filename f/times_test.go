package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
	"time"
)

func TestAt_Future(test *testing.T) {
	start := time.Now()
	done := f.At(start.Add(100*time.Millisecond), func() {
		diff := time.Now().Sub(start)
		if diff > 105*time.Millisecond {
			test.Errorf("Expected to run in 100 ms, it did in %v", diff)
		}
	})
	<-done
}

func TestAt_Past(test *testing.T) {
	start := time.Now()
	done := f.At(start.Add(-100*time.Millisecond), func() {})
	<-done
	diff := time.Now().Sub(start)
	if diff > time.Millisecond {
		test.Errorf("Expected to return immediately, but it took %v", diff)
	}
}

func TestAfter_Future(test *testing.T) {
	start := time.Now()
	done := f.After(100*time.Millisecond, func() {
		diff := time.Now().Sub(start)
		if diff > 105*time.Millisecond {
			test.Errorf("Expected to run in 100 ms, it did in %v", diff)
		}
	})
	<-done
}

func TestEvery(test *testing.T) {
	dur := 10 * time.Millisecond
	count := 0
	f.Every(dur, func() {
		count++
	})
	<-time.After(100 * time.Millisecond)
	count++
	if count < 9 {
		test.Errorf("Expected to run in at least 9 times, it did %v times", count)
	}
}

func TestUntil_Future(test *testing.T) {
	count := 0
	done := f.Until(time.Now().Add(100*time.Millisecond), 10*time.Millisecond, func() {
		count++
	})
	<-done
	if count < 9 {
		test.Errorf("Expected to run for at least for 9 times, but it ran for %v times", count)
	}
}

func TestUntil_Past(test *testing.T) {
	count := 0
	done := f.Until(time.Now().Add(-100*time.Millisecond), 10*time.Millisecond, func() {
		count++
	})
	<-done
	if count != 0 {
		test.Errorf("Expected to run for at least for 0 times, but it ran for %v times", count)
	}
}

func TestTimeout_TimedOut(test *testing.T) {
	done := f.Timeout(100*time.Millisecond, func() {
		time.Sleep(time.Minute)
	})
	if <-done {
		test.Errorf("Expected to get timed out, but it has been completed")
	}
}

func TestTimeout_Completed(test *testing.T) {
	done := f.Timeout(time.Minute, func() {
		time.Sleep(100 * time.Millisecond)
	})
	if !<-done {
		test.Errorf("Expected to get completed, but it has been timed out")
	}
}

func TestAll(test *testing.T) {
	start := time.Now()
	var val1, val2, val3 bool
	done := f.All(func() {
		val1 = true
		time.Sleep(100 * time.Millisecond)
	}, func() {
		val2 = true
		time.Sleep(100 * time.Millisecond)
	}, func() {
		val3 = true
		time.Sleep(100 * time.Millisecond)
	})
	<-done
	diff := time.Now().Sub(start)
	if diff > 105*time.Millisecond {
		test.Errorf("All takes too long to complete")
	}
	if !(val1 && val2 && val3) {
		test.Errorf("Expected all to run, but at least one didn't")
	}
}

func TestAllWithThrottle(test *testing.T) {
	start := time.Now()
	fn := func() {
		time.Sleep(100 * time.Millisecond)
	}
	done := f.AllWithThrottle(3, fn, fn, fn, fn, fn)
	<-done
	diff := time.Now().Sub(start)
	if diff > 205*time.Millisecond {
		test.Errorf("All with throttle takes too long to complete")
	}
	if diff < 105*time.Millisecond {
		test.Errorf("All with throttle doesn't take long, throttling may not work")
	}
}

func TestReplicate(test *testing.T) {
	results := make(chan bool, 5)
	done := f.Replicate(5, func() {
		results <- true
	})
	<-done
	close(results)
	count := 0
	for _ = range results {
		count++
	}
	if count != 5 {
		test.Errorf("Expected 5 to run, but %v worked", count)
	}
}
