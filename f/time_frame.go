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
		Since: TimeFrom(since, true),
		Until: TimeFrom(until, true),
		Data:  make([]byte, 0),
	}
	return t
}

func NewTimeFrames(since, until time.Time, duration time.Duration) []*TimeFrame {
	since, until = time.Unix(since.Unix(), 0).Local(), time.Unix(until.Unix(), 0).Local()
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
	return s
}

func (t *TimeFrame) Now() bool {
	u := time.Now().Unix()
	return t.Since.UnixSecondTimeStamp <= u && u < t.Until.UnixSecondTimeStamp
}
