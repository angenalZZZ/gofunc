package f

import (
	"strconv"
	"strings"
	"time"
)

const (
	DateFormatString     string = "2006-01-02"
	DateTimeFormatString string = "2006-01-02 15:04:05"
	TimeFormatString     string = "2006-01-02 15:04:05.000"
)

// TimeStamp a time stamp.
type TimeStamp int64

// String a String method.
func (t TimeStamp) String() string {
	return strconv.FormatInt(int64(t), 10)
}

// TimeStampObject get unix/UTC time stamp,
// 精确到秒 10位数: 1582950407  the number of seconds elapsed since January 1, 1970 UTC.
func TimeStampObject() TimeStamp {
	return TimeStamp(time.Now().Unix())
}

// UTCTimeStampString get UTC time string,
// 精确到秒 14位数: 20200229042647  the number of second.
func (t TimeStamp) UTCTimeStampString() string {
	return t.AsUTCTime().Format("20060102150405")
}

// LocalTimeStampString get Local time string,
// 精确到秒 14位数: 20200229122647  the number of second.
func (t TimeStamp) LocalTimeStampString() string {
	return t.AsLocalTime().Format("20060102150405")
}

// LocalTimeStampString get Local time string,
// 精确到毫秒 17位数: 20200229122647003  the number of millisecond.
func LocalTimeStampString() string {
	t := time.Now().Format("20060102150405.000")
	return strings.Replace(t, ".", "", 1)
}

// LocalTimeString get Local time string,
// 精确到毫秒: 2020-02-29 12:26:47.003  the time of millisecond.
func LocalTimeString() string {
	return time.Now().Format(TimeFormatString)
}

// UTCTimeString get UTC time string,
// 精确到秒: 2020-02-29 04:26:47  the time of second.
func (t TimeStamp) UTCTimeString() string {
	return t.AsUTCTime().Format(DateTimeFormatString)
}

// LocalTimeString get Local time string,
// 精确到秒: 2020-02-29 12:26:47  the time of second.
func (t TimeStamp) LocalTimeString() string {
	return t.AsLocalTime().Format(DateTimeFormatString)
}

// LocalDateString get Local date string,
// 精确到天: 2020-02-29  the date.
func (t TimeStamp) LocalDateString() string {
	return t.AsLocalTime().Format(DateFormatString)
}

// FormatLocalTimeString formats Local time string.
func (t TimeStamp) FormatLocalTimeString(f string) string {
	return t.AsLocalTime().Format(f)
}

// AsLocalTime Convert timestamp as time in Local locale.
func (t TimeStamp) AsLocalTime() time.Time {
	return time.Unix(int64(t), 0).In(time.Local)
}

// AsUTCTime Convert timestamp as time in UTC locale.
func (t TimeStamp) AsUTCTime() time.Time {
	return time.Unix(int64(t), 0).In(time.UTC)
}

// AddSeconds adds seconds and return sum.
func (t TimeStamp) AddSeconds(seconds int64) TimeStamp {
	return t + TimeStamp(seconds)
}

// AddDuration adds time.Duration and return sum.
func (t TimeStamp) AddDuration(interval time.Duration) TimeStamp {
	return t + TimeStamp(interval/time.Second)
}

// IsZero is zero time.
func (t TimeStamp) IsZero() bool {
	return t.AsLocalTime().IsZero()
}

// Values returns the time's year, month, day, hour, minute, second.
func (t TimeStamp) Values() (year, month, day, hour, minute, second int) {
	i := t.AsLocalTime()
	return i.Year(), int(i.Month()), i.Day(), i.Hour(), i.Minute(), i.Second()
}
