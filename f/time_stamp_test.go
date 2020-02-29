package f

import "testing"

func TestTimeStamp(t *testing.T) {
	ts := Now()
	t.Log(ts)
	t.Log(ts.UnixSecondTimeStampString())
	t.Log(ts.MilliSecondTimeStampString())
	t.Log(ts.MicroSecondTimeStampString())
	t.Log(ts.NanoSecondTimeStampString())
	t.Log(ts.UTCTimeStampString())
	t.Log(ts.LocalTimeStampString())
	t.Log(ts.UTCString())
	t.Log(ts.LocalString())
	t.Log(ts.UTCTimeString())
	t.Log(ts.LocalTimeString())
	t.Log(ts.UTCDateString())
	t.Log(ts.LocalDateString())
}
