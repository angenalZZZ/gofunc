package f

import (
	"time"
)

type TimeFrame struct {
	Since, Until *TimeStamp
	Data         []byte
}

func NewTimeFrame(since, until time.Time) *TimeFrame {
	t := &TimeFrame{
		Since: TimeFrom(since),
		Until: TimeFrom(until),
		Data:  make([]byte, 0),
	}
	return t
}

func NewTimeFrames(since, until time.Time, duration time.Duration) []*TimeFrame {
	a, s := since, make([]*TimeFrame, 0)
	for a.Before(until) {
		// set to date = from + duration
		t := a.Add(duration)

		if t.Before(until) {
			s = append(s, NewTimeFrame(a, t))
		} else {
			s = append(s, NewTimeFrame(a, until))
		}

		// increment from date with 1
		a = t
	}
}

func (t *TimeFrame) Now() bool {
	u := time.Now()
	return t.Since.Time.Before(u) && u.Before(t.Until.Time)
}
