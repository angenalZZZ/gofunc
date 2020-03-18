package f

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	DateFormatStringG     string = "20060102"
	DateFormatString      string = "2006-01-02"
	DateTimeFormatStringH string = "2006-01-02 15"
	DateTimeFormatStringM string = "2006-01-02 15:04"
	DateTimeFormatString  string = "2006-01-02 15:04:05"
	TimeFormatString      string = "2006-01-02 15:04:05.000"
)

// TimeStamp a time stamp and extended methods.
type TimeStamp struct {
	time.Time
}

// UnixSecondTimeStampString 时间戳 unix time stamp,
// 精确到秒 10位数: 1582950407
// the number of seconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) UnixSecondTimeStampString() string {
	return strconv.FormatInt(t.Time.Unix(), 10)
}

// MilliSecondTimeStampString 时间戳 unix time stamp,
// 精确到毫秒 13位数: 1582950407018
// the number of milliseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) MilliSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%03d", t.Time.Nanosecond()/1e6)
}

// MicroSecondTimeStampString 时间戳 unix time stamp,
// 精确到微秒 16位数: 1582950407018018
// the number of microseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) MicroSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%06d", t.Time.Nanosecond()/1e3)
}

// NanoSecondTimeStampString 时间戳 unix time stamp,
// 精确到纳秒 19位数: 1582950407018018100
// the number of nanoseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) NanoSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%09d", t.Time.Nanosecond())
}

// Now get now timestamp in Local time.
func Now() *TimeStamp {
	return TimeFrom(time.Now())
}

// TimeFrom get a timestamp in Local time.
func TimeFrom(t time.Time) *TimeStamp {
	ts := &TimeStamp{t}
	return ts
}

// TimeStampFrom get a timestamp in Local time.
// the timestamp length equals(10/13/16/19) since January 1, 1970 UTC.
func TimeStampFrom(timestamp string) *TimeStamp {
	if len(timestamp) < 10 {
		return nil
	}
	seconds, err := strconv.ParseInt(timestamp[0:10], 10, 64)
	if err != nil {
		return nil
	}
	nanoSeconds := 0
	switch n := timestamp[10:]; len(n) {
	case 3:
		if i, err := strconv.Atoi(n); err != nil {
			return nil
		} else {
			nanoSeconds = i * 1e6
		}
	case 6:
		if i, err := strconv.Atoi(n); err != nil {
			return nil
		} else {
			nanoSeconds = i * 1e3
		}
	case 9:
		if i, err := strconv.Atoi(n); err != nil {
			return nil
		} else {
			nanoSeconds = i
		}
	default:
		return nil
	}
	return TimeStampFromSeconds(seconds, int64(nanoSeconds))
}

// TimeStampFromSeconds get a timestamp in Local time.
// the number of seconds and nanoSeconds since January 1, 1970 UTC.
func TimeStampFromSeconds(seconds int64, nanoSeconds int64) *TimeStamp {
	t := time.Unix(seconds, nanoSeconds).In(time.Local)
	ts := &TimeStamp{t}
	return ts
}

// UTCTimeStampString get UTC time string,
// 精确到毫秒 17位数: 20200202042647003  the number of second.
func (t *TimeStamp) UTCTimeStampString() string {
	s := t.AsUTCTime().Format("20060102150405.000")
	return strings.Replace(s, ".", "", 1)
}

// LocalTimeStampString get Local time string,
// 精确到毫秒 17位数: 20200202122647003  the number of second.
func (t *TimeStamp) LocalTimeStampString() string {
	s := t.Time.Format("20060102150405.000")
	return strings.Replace(s, ".", "", 1)
}

// UTCString get UTC time string,
// 精确到秒: 2020-02-02 04:26:47  the time of second.
func (t *TimeStamp) UTCString() string {
	return t.AsUTCTime().Format(DateTimeFormatString)
}

// LocalString get Local time string,
// 精确到秒: 2020-02-02 12:26:47  the time of second.
func (t *TimeStamp) LocalString() string {
	return t.Time.Format(DateTimeFormatString)
}

// UTCTimeString get UTC time string,
// 精确到毫秒: 2020-02-02 04:26:47.003  the time of millisecond.
func (t *TimeStamp) UTCTimeString() string {
	return t.AsUTCTime().Format(TimeFormatString)
}

// LocalTimeString get Local time string,
// 精确到毫秒: 2020-02-02 12:26:47.003  the time of millisecond.
func (t *TimeStamp) LocalTimeString() string {
	return t.Time.Format(TimeFormatString)
}

// UTCDateString get UTC date string,
// 精确到天: 2020-02-02  the date.
func (t *TimeStamp) UTCDateString() string {
	return t.AsUTCTime().Format(DateFormatString)
}

// LocalDateString get Local date string,
// 精确到天: 2020-02-02  the date.
func (t *TimeStamp) LocalDateString() string {
	return t.Time.Format(DateFormatString)
}

// AsTime get a time in Local locale.
func (t *TimeStamp) AsTime() time.Time {
	return t.Time
}

// AsTimeIn Convert timestamp as time in a locale, equals t.In(local).
func (t *TimeStamp) AsTimeIn(local *time.Location) time.Time {
	return time.Unix(t.Unix(), int64(t.Nanosecond())).In(local)
}

// AsLocalTime Convert timestamp as time in Local locale.
func (t *TimeStamp) AsLocalTime() time.Time {
	return t.Time.Local()
}

// AsUTCTime Convert timestamp as time in UTC locale.
func (t *TimeStamp) AsUTCTime() time.Time {
	return t.Time.UTC()
}

// AddSeconds adds seconds and return sum.
func (t *TimeStamp) AddSeconds(seconds int64) *TimeStamp {
	t.Time.Add(time.Duration(seconds) * time.Second)
	return t
}

// AddMinutes adds minutes and return sum.
func (t *TimeStamp) AddMinutes(minutes int64) *TimeStamp {
	t.Time.Add(time.Duration(minutes) * time.Minute)
	return t
}

// AddHours adds hours and return sum.
func (t *TimeStamp) AddHours(hours int64) *TimeStamp {
	t.Time.Add(time.Duration(hours) * time.Hour)
	return t
}

// YearMonthDay returns the time's year, month, day.
func (t *TimeStamp) YearMonthDay() (year, month, day int) {
	return t.Time.Year(), int(t.Time.Month()), t.Time.Day()
}

// HourMinuteSecond returns the time's hour, minute, second.
func (t *TimeStamp) HourMinuteSecond() (hour, minute, second int) {
	return t.Time.Hour(), t.Time.Minute(), t.Time.Second()
}

// ToTime convert string to time.Time
func ToTime(s string, layouts ...string) (t time.Time, err error) {
	layout := toTimeLayout(s, layouts...)
	if layout == "" {
		err = errConvertFail
		return
	}
	t, err = time.Parse(layout, s)
	return
}

// ToLocalTime convert string to time.Time
func ToLocalTime(s string, layouts ...string) (t time.Time, err error) {
	layout := toTimeLayout(s, layouts...)
	if layout == "" {
		err = errConvertFail
		return
	}
	t, err = time.ParseInLocation(layout, s, time.Local)
	return
}

func toTimeLayout(s string, layouts ...string) string {
	var layout string
	if len(layouts) > 0 {
		layout = layouts[0]
	} else {
		switch len(s) {
		case 8:
			layout = DateFormatStringG
		case 10:
			layout = DateFormatString
		case 13:
			layout = DateTimeFormatStringH
		case 16:
			layout = DateTimeFormatStringM
		case 19:
			layout = DateTimeFormatString
		case 20:
		case 25:
			layout = time.RFC3339
		case 30:
		case 35:
			layout = time.RFC3339Nano
		}
	}

	if layout != "" {
		// has 'T' eg.2006-01-02T15:04:05
		if strings.ContainsRune(s, 'T') {
			layout = strings.Replace(layout, " ", "T", -1)
		}
		// eg: 2006/01/02 15:04:05
		if strings.ContainsRune(s, '/') {
			layout = strings.Replace(layout, "-", "/", -1)
		}
	}

	return layout
}

// IsDate check value is an date string.
func IsDate(srcDate string) bool {
	_, err := ToTime(srcDate)
	return err == nil
}

// BeforeDate check
func BeforeDate(srcDate, dstDate string) bool {
	st, err := ToTime(srcDate)
	if err != nil {
		return false
	}

	dt, err := ToTime(dstDate)
	if err != nil {
		return false
	}

	return st.Before(dt)
}

// BeforeOrEqualDate check
func BeforeOrEqualDate(srcDate, dstDate string) bool {
	st, err := ToTime(srcDate)
	if err != nil {
		return false
	}

	dt, err := ToTime(dstDate)
	if err != nil {
		return false
	}

	return st.Before(dt) || st.Equal(dt)
}

// AfterOrEqualDate check
func AfterOrEqualDate(srcDate, dstDate string) bool {
	st, err := ToTime(srcDate)
	if err != nil {
		return false
	}

	dt, err := ToTime(dstDate)
	if err != nil {
		return false
	}

	return st.After(dt) || st.Equal(dt)
}

// AfterDate check
func AfterDate(srcDate, dstDate string) bool {
	st, err := ToTime(srcDate)
	if err != nil {
		return false
	}

	dt, err := ToTime(dstDate)
	if err != nil {
		return false
	}

	return st.After(dt)
}
