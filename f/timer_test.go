package f_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/angenalZZZ/gofunc/f"
)

func ExampleTimer_Start() {
	tw := f.NewTimer(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	exitC := make(chan time.Time, 1)
	tw.AfterFunc(time.Second, func() {
		fmt.Println("The timer fires")
		exitC <- time.Now().UTC()
	})

	<-exitC

	// Output:
	// The timer fires
}

func ExampleTimer_Stop() {
	tw := f.NewTimer(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	t := tw.AfterFunc(time.Second, func() {
		fmt.Println("The timer fires")
	})

	<-time.After(900 * time.Millisecond)
	// Stop the timer before it fires
	t.Stop()

	// Output:
	//
}

func TestTimerBucket_Flush(t *testing.T) {
	b := f.NewTimerBucket()

	b.Add(&f.TimerElement{})
	b.Add(&f.TimerElement{})
	l1 := b.Timers.Len()
	if l1 != 2 {
		t.Fatalf("Got (%+v) != Want (%+v)", l1, 2)
	}

	b.Flush(func(*f.TimerElement) {})
	l2 := b.Timers.Len()
	if l2 != 0 {
		t.Fatalf("Got (%+v) != Want (%+v)", l2, 0)
	}
}

func TestTimer_AfterFunc(t *testing.T) {
	tw := f.NewTimer(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	durations := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
	}
	for _, d := range durations {
		t.Run("", func(t *testing.T) {
			exitC := make(chan time.Time)

			start := time.Now().UTC()
			tw.AfterFunc(d, func() {
				exitC <- time.Now().UTC()
			})

			got := (<-exitC).Truncate(time.Millisecond)
			min := start.Add(d).Truncate(time.Millisecond)

			err := 5 * time.Millisecond
			if got.Before(min) || got.After(min.Add(err)) {
				t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", d, min, min.Add(err), got)
			}
		})
	}
}

type anTimerScheduler struct {
	intervals []time.Duration
	current   int
}

func (s *anTimerScheduler) Next(prev time.Time) time.Time {
	if s.current >= len(s.intervals) {
		return time.Time{}
	}
	next := prev.Add(s.intervals[s.current])
	s.current++
	return next
}

func TestTimer_ScheduleFunc(t *testing.T) {
	tw := f.NewTimer(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	s := &anTimerScheduler{intervals: []time.Duration{
		1 * time.Millisecond,   // start + 1ms
		4 * time.Millisecond,   // start + 5ms
		5 * time.Millisecond,   // start + 10ms
		40 * time.Millisecond,  // start + 50ms
		50 * time.Millisecond,  // start + 100ms
		400 * time.Millisecond, // start + 500ms
		500 * time.Millisecond, // start + 1s
	}}

	exitC := make(chan time.Time, len(s.intervals))

	start := time.Now().UTC()
	tw.ScheduleFunc(s, func() {
		exitC <- time.Now().UTC()
	})

	accum := time.Duration(0)
	for _, d := range s.intervals {
		got := (<-exitC).Truncate(time.Millisecond)
		accum += d
		min := start.Add(accum).Truncate(time.Millisecond)

		err := 5 * time.Millisecond
		if got.Before(min) || got.After(min.Add(err)) {
			t.Errorf("Timer(%s) expiration: want [%s, %s], got %s", accum, min, min.Add(err), got)
		}
	}
}
