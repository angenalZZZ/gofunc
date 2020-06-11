package f

import (
	"time"
)

// TimeFrame - [Since ~ Until) eg. [1, 15) included 1, not included 15
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
		t := a.Add(duration)
		if t.Before(until) {
			s = append(s, NewTimeFrame(a, t))
		} else {
			s = append(s, NewTimeFrame(a, until))
		}
		a = t
	}
	return s
}

func (t *TimeFrame) In(t2 time.Time) bool {
	u := t2.Unix()
	return t.Since.UnixSecondTimeStamp <= u && u < t.Until.UnixSecondTimeStamp
}

func (t *TimeFrame) String() string {
	p, _ := EncodeJson(map[string]string{
		"Since": t.Since.LocalString(),
		"Until": t.Until.LocalString(),
		"Data":  string(t.Data),
	})
	return String(p)
}
