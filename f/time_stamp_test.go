package f

import (
	"testing"
	"time"
)

func TestTimeStamp(t *testing.T) {
	ts := Now() // equals f.TimeFrom(time.Now())
	ts = TimeFrom(time.Now())
	t.Log(ts.LocalString())
	t.Log(NewTimeStamp(0))       // 1970-01-01 08:00:00 +0800 CST
	t.Log(NewTimeStamp(0).UTC()) // 1970-01-01 00:00:00 +0000 UTC
	ts, _ = TimeFromLocalString("2020-03-08 11:19:42")
	t.Log(ts.UTCTimeString(), ts.LocalTimeString())
	ts, _ = TimeFromUTCString("2020-03-08 03:19:42")
	t.Log(ts.UTCTimeString(), ts.LocalTimeString())
	ts = TimeStampFrom("1583637582")
	ts = TimeStampFrom("1583637582780")
	ts = TimeStampFrom("1583637582780102")
	ts = TimeStampFrom("1583637582780102300")
	t.Log(ts.String())                     // Output: 2020-03-08 11:19:42.7801023 +0800 CST
	t.Log(ts.UnixSecondTimeStampString())  // Output: 1583637582
	t.Log(ts.UnixSecondTimeStamp())        // Output: 1583637582
	t.Log(ts.MilliSecondTimeStampString()) // Output: 1583637582780
	t.Log(ts.MilliSecondTimeStamp())       // Output: 1583637582780
	t.Log(ts.MicroSecondTimeStampString()) // Output: 1583637582780102
	t.Log(ts.MicroSecondTimeStamp())       // Output: 1583637582780102
	t.Log(ts.NanoSecondTimeStampString())  // Output: 1583637582780102300
	t.Log(ts.NanoSecondTimeStamp())        // Output: 1583637582780102300
	t.Log(ts.UTCTimeStampString())         // Output: 20200308031942780
	t.Log(ts.LocalTimeStampString())       // Output: 20200308111942780
	t.Log(ts.UTCString())                  // Output: 2020-03-08 03:19:42
	t.Log(ts.LocalString())                // Output: 2020-03-08 11:19:42
	t.Log(ts.UTCTimeString())              // Output: 2020-03-08 03:19:42.780
	t.Log(ts.LocalTimeString())            // Output: 2020-03-08 11:19:42.780
	t.Log(ts.UTCDateString())              // Output: 2020-03-08
	t.Log(ts.LocalDateString())            // Output: 2020-03-08
}
