package f_test

import (
	"github.com/angenalZZZ/gofunc/f"
	"testing"
	"time"
)

func TestTimeStamp(t *testing.T) {
	ts := f.Now() // equals f.TimeFrom(time.Now())
	ts = f.TimeFrom(time.Now())
	t.Log(ts.LocalString())
	t.Log(f.TimeFrom(ts.Time), f.TimeFrom(ts.Time, true))
	t.Log(f.NewTimeStamp(0))       // 1970-01-01 08:00:00 +0800 CST
	t.Log(f.NewTimeStamp(0).UTC()) // 1970-01-01 00:00:00 +0000 UTC
	ts, _ = f.TimeFromLocalString("2020-03-08 11:19:42")
	ts, _ = f.TimeFromLocalString("20200308111942000")
	t.Log(ts.UTCTimeString(), ts.LocalTimeString())
	ts, _ = f.TimeFromUTCString("2020-03-08 03:19:42")
	t.Log(ts.UTCTimeString(), ts.LocalTimeString())
	ts = f.TimeStampFrom("1583637582")
	ts = f.TimeStampFrom("1583637582780")
	ts = f.TimeStampFrom("1583637582780102")
	ts = f.TimeStampFrom("1583637582780102300")
	t.Log(ts.String())                     // Output: 2020-03-08 11:19:42.7801023 +0800 CST
	t.Log(ts.UnixSecondTimeStampString())  // Output: 1583637582
	t.Log(ts.UnixSecond)                   // Output: 1583637582
	t.Log(ts.MilliSecondTimeStampString()) // Output: 1583637582780
	t.Log(ts.MilliSecondTimeStamp())       // Output: 1583637582780
	t.Log(ts.MicroSecondTimeStampString()) // Output: 1583637582780102
	t.Log(ts.MicroSecondTimeStamp())       // Output: 1583637582780102
	t.Log(ts.NanoSecondTimeStampString())  // Output: 1583637582780102300
	t.Log(ts.UnixNanoSecond)               // Output: 1583637582780102300
	t.Log(ts.UTCTimeStampString())         // Output: 20200308031942780
	t.Log(ts.LocalTimeStampString())       // Output: 20200308111942780
	t.Log(ts.UTCString())                  // Output: 2020-03-08 03:19:42
	t.Log(ts.LocalString())                // Output: 2020-03-08 11:19:42
	t.Log(ts.UTCTimeString())              // Output: 2020-03-08 03:19:42.780
	t.Log(ts.LocalTimeString())            // Output: 2020-03-08 11:19:42.780
	t.Log(ts.UTCDateString())              // Output: 2020-03-08
	t.Log(ts.LocalDateString())            // Output: 2020-03-08
}

func TestIsTime(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		param    string
		format   string
		expected bool
	}{
		{"2016-12-31 11:00", time.RFC3339, false},
		{"2016-12-31 11:00:00", time.RFC3339, false},
		{"2016-12-31T11:00", time.RFC3339, false},
		{"2016-12-31T11:00:00", time.RFC3339, false},
		{"2016-12-31T11:00:00Z", time.RFC3339, true},
		{"2016-12-31T11:00:00+01:00", time.RFC3339, true},
		{"2016-12-31T11:00:00-01:00", time.RFC3339, true},
		{"2016-12-31T11:00:00.05Z", time.RFC3339, true},
		{"2016-12-31T11:00:00.05-01:00", time.RFC3339, true},
		{"2016-12-31T11:00:00.05+01:00", time.RFC3339, true},
		{"2016-12-31T11:00:00", f.RF3339WithoutZone, true},
		{"2016-12-31T11:00:00Z", f.RF3339WithoutZone, false},
		{"2016-12-31T11:00:00+01:00", f.RF3339WithoutZone, false},
		{"2016-12-31T11:00:00-01:00", f.RF3339WithoutZone, false},
		{"2016-12-31T11:00:00.05Z", f.RF3339WithoutZone, false},
		{"2016-12-31T11:00:00.05-01:00", f.RF3339WithoutZone, false},
		{"2016-12-31T11:00:00.05+01:00", f.RF3339WithoutZone, false},
	}
	for _, test := range tests {
		actual := f.IsTime(test.param, test.format)
		if actual != test.expected {
			t.Errorf("Expected IsTime(%q, time.RFC3339) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}

func TestIsRFC3339(t *testing.T) {
	t.Parallel()
	var tests = []struct {
		param    string
		expected bool
	}{
		{"2016-12-31 11:00", false},
		{"2016-12-31 11:00:00", false},
		{"2016-12-31T11:00", false},
		{"2016-12-31T11:00:00", false},
		{"2016-12-31T11:00:00Z", true},
		{"2016-12-31T11:00:00+01:00", true},
		{"2016-12-31T11:00:00-01:00", true},
		{"2016-12-31T11:00:00.05Z", true},
		{"2016-12-31T11:00:00.05-01:00", true},
		{"2016-12-31T11:00:00.05+01:00", true},
	}
	for _, test := range tests {
		actual := f.IsRFC3339(test.param)
		if actual != test.expected {
			t.Errorf("Expected IsRFC3339(%q) to be %v, got %v", test.param, test.expected, actual)
		}
	}
}
