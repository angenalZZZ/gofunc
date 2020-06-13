package f

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	ZoneName         string
	ZoneOffset       int
	ZoneOffsetSecond time.Duration
)

func init() {
	ZoneName, ZoneOffset = time.Now().Zone()
	ZoneOffsetSecond = time.Duration(ZoneOffset) * time.Second
}

// TimeStamp a time stamp and extended methods.
type TimeStamp struct {
	time.Time
	UnixSecond     int64
	UnixNanoSecond int64
}

// UnixSecondTimeStampString 时间戳 unix time stamp,
// 精确到秒 10位数: 1582950407
// the number of seconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) UnixSecondTimeStampString() string {
	return strconv.FormatInt(t.UnixSecond, 10)
}

// MilliSecondTimeStampString 时间戳 unix time stamp,
// 精确到毫秒 13位数: 1582950407018
// the number of milliseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) MilliSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%03d", t.Time.Nanosecond()/1e6)
}

func (t *TimeStamp) MilliSecondTimeStamp() int64 {
	return t.UnixNanoSecond / 1e6
}

// MicroSecondTimeStampString 时间戳 unix time stamp,
// 精确到微秒 16位数: 1582950407018018
// the number of microseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) MicroSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%06d", t.Time.Nanosecond()/1e3)
}

func (t *TimeStamp) MicroSecondTimeStamp() int64 {
	return t.UnixNanoSecond / 1e3
}

// NanoSecondTimeStampString 时间戳 unix time stamp,
// 精确到纳秒 19位数: 1582950407018018100
// the number of nanoseconds elapsed since January 1, 1970 UTC. The result does not depend on the
// location associated with t.
func (t *TimeStamp) NanoSecondTimeStampString() string {
	return t.UnixSecondTimeStampString() + fmt.Sprintf("%09d", t.Time.Nanosecond())
}

// Now get now timestamp in Local time.
// upToSecond is used to remove milliseconds.
func Now(upToSecond ...bool) *TimeStamp {
	return TimeFrom(time.Now(), upToSecond...)
}

// NowLocalString get a time at now.
func NowLocalString(upToSecond ...bool) string {
	return Now(upToSecond...).LocalString()
}

// TimeFrom get a timestamp in Local time.
// upToSecond is used to remove milliseconds.
func TimeFrom(t time.Time, upToSecond ...bool) *TimeStamp {
	ts := &TimeStamp{t, t.Unix(), 0}
	if len(upToSecond) > 0 && upToSecond[0] == true {
		ts.Time = time.Unix(ts.UnixSecond, 0).Local()
		ts.UnixNanoSecond = ts.UnixSecond * 1e9
	} else {
		ts.UnixNanoSecond = t.UnixNano()
	}
	return ts
}

// TimeFromLocalString get a timestamp in Time string layouts.
func TimeFromLocalString(s string, layouts ...string) (*TimeStamp, error) {
	if t, err := ToUTCTime(s, layouts...); err != nil {
		return nil, err
	} else {
		t = t.Local()
		return &TimeStamp{t, t.Unix(), t.UnixNano()}, nil
	}
}

// TimeFromUTCString get a timestamp in Time string layouts.
func TimeFromUTCString(s string, layouts ...string) (*TimeStamp, error) {
	if t, err := ToTime(s, layouts...); err != nil {
		return nil, err
	} else {
		t = t.Local()
		return &TimeStamp{t, t.Unix(), t.UnixNano()}, nil
	}
}

// TimeFrom get a timestamp in Time bytes.
func TimeFromBytes(data []byte) (*TimeStamp, error) {
	t := time.Time{}
	if err := t.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	ts := &TimeStamp{t, t.Unix(), t.UnixNano()}
	return ts, nil
}

// NewTimeStamp convert a timestamp to Local time.
// the timestamp length equals(10/13/16/19) since January 1, 1970 UTC.
func NewTimeStamp(i int64) *TimeStamp {
	t, _ := time.Parse(DateTimeFormatString, "1970-01-01 00:00:00")
	if i < 1e12 {
		t = t.Add(time.Duration(i) * time.Second)
	} else if i < 1e15 {
		t = t.Add(time.Duration(i/1e3) * time.Second).Add(time.Duration(i%1e3) * time.Millisecond)
	} else if i < 1e18 {
		t = t.Add(time.Duration(i/1e6) * time.Second).Add(time.Duration(i%1e6) * time.Microsecond)
	} else {
		t = t.Add(time.Duration(i/1e9) * time.Second).Add(time.Duration(i%1e9) * time.Nanosecond)
	}
	t = t.Local()
	ts := &TimeStamp{t, t.Unix(), t.UnixNano()}
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
	t := time.Unix(seconds, nanoSeconds).Local()
	ts := &TimeStamp{t, t.Unix(), t.UnixNano()}
	return ts
}

// UTCTimeStampString get UTC time string,
// 精确到毫秒 17位数: 20200202042647003
// 或精确到秒 14位数: 20200202042647 (upToSecond=true)
func (t *TimeStamp) UTCTimeStampString(upToSecond ...bool) string {
	if len(upToSecond) > 0 && upToSecond[0] == true {
		return t.AsUTCTime().Format(TimeFormatStringS)
	}
	s := t.AsUTCTime().Format(TimeFormatStringM)
	return strings.Replace(s, ".", "", 1)
}

// LocalTimeStampString get Local time string,
// 精确到毫秒 17位数: 20200202122647003
// 或精确到秒 14位数: 20200202042647 (upToSecond=true)
func (t *TimeStamp) LocalTimeStampString(upToSecond ...bool) string {
	if len(upToSecond) > 0 && upToSecond[0] == true {
		return t.Time.Format(TimeFormatStringS)
	}
	s := t.Time.Format(TimeFormatStringM)
	return strings.Replace(s, ".", "", 1)
}

// UTCString get UTC time string,
// 精确到秒: 2020-02-02 04:26:47  the time of second.
func (t *TimeStamp) UTCString(layouts ...string) string {
	layout := DateTimeFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	if name, _ := t.Time.Zone(); name == ZoneName {
		return toUTCTime(t.Time).Format(layout)
	}
	return t.AsUTCTime().Format(layout)
}

// LocalString get Local time string,
// 精确到秒: 2020-02-02 12:26:47  the time of second.
func (t *TimeStamp) LocalString(layouts ...string) string {
	layout := DateTimeFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	if name, _ := t.Time.Zone(); name != ZoneName {
		return toLocalTime(t.Time).Format(layout)
	}
	return t.Time.Format(layout)
}

// UTCTimeString get UTC time string,
// 精确到毫秒: 2020-02-02 04:26:47.003  the time of millisecond.
func (t *TimeStamp) UTCTimeString(layouts ...string) string {
	layout := TimeFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	if name, _ := t.Time.Zone(); name == ZoneName {
		return toUTCTime(t.Time).Format(layout)
	}
	return t.AsUTCTime().Format(layout)
}

// LocalTimeString get Local time string,
// 精确到毫秒: 2020-02-02 12:26:47.003  the time of millisecond.
func (t *TimeStamp) LocalTimeString(layouts ...string) string {
	layout := TimeFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	if name, _ := t.Time.Zone(); name != ZoneName {
		return toLocalTime(t.Time).Format(layout)
	}
	return t.Time.Format(layout)
}

// UTCDateString get UTC date string,
// 精确到天: 2020-02-02  the date.
func (t *TimeStamp) UTCDateString(layouts ...string) string {
	layout := DateFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	return t.AsUTCTime().Format(layout)
}

// LocalDateString get Local date string,
// 精确到天: 2020-02-02  the date.
func (t *TimeStamp) LocalDateString(layouts ...string) string {
	layout := DateFormatString
	if len(layouts) > 0 {
		layout = layouts[0]
	}
	return t.Time.Format(layout)
}

// AsTime get a time in Local locale.
func (t *TimeStamp) AsTime() time.Time {
	return t.Time
}

// AsTimeIn Convert timestamp as time in a locale, equals t.In(local).
func (t *TimeStamp) AsTimeIn(local *time.Location) time.Time {
	return time.Unix(t.UnixSecond, int64(t.Nanosecond())).In(local)
}

// AsLocal Convert timestamp as time for Local locale.
func (t *TimeStamp) AsLocal() *TimeStamp {
	t.Time = toLocalTime(t.Time)
	return t
}

// AsLocalTime Convert timestamp as time for Local locale.
func (t *TimeStamp) AsLocalTime() time.Time {
	return t.Time.Local()
}

// AsUTC Convert timestamp as time for UTC locale.
func (t *TimeStamp) AsUTC() *TimeStamp {
	t.Time = toUTCTime(t.Time)
	return t
}

// AsUTCTime Convert timestamp as time for UTC locale.
func (t *TimeStamp) AsUTCTime() time.Time {
	return t.Time.UTC()
}

// ToLocalTime Convert timestamp as time in Local locale, add +8 hours.
func (t *TimeStamp) ToLocalTime() time.Time {
	t.Time = toLocalTime(t.Time)
	return t.Time.Local()
}

// ToUTCTime Convert timestamp as time in UTC locale, add -8 hours.
func (t *TimeStamp) ToUTCTime() time.Time {
	t.Time = toUTCTime(t.Time)
	return t.Time.UTC()
}

// ToLocal Convert timestamp as time in Local locale, add +8 hours.
func (t *TimeStamp) ToLocal() *TimeStamp {
	t.Time = t.ToLocalTime()
	return t
}

// ToUTC Convert timestamp as time in UTC locale, add -8 hours.
func (t *TimeStamp) ToUTC() *TimeStamp {
	t.Time = t.ToUTCTime()
	return t
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

// ToBytes Time MarshalBinary.
func (t *TimeStamp) ToBytes() []byte {
	data, _ := t.Time.MarshalBinary()
	return data
}

// ToJSON Time MarshalJSON.
func (t *TimeStamp) ToJSON() []byte {
	data, _ := t.Time.MarshalJSON()
	return data
}

// ToText Time MarshalText.
func (t *TimeStamp) ToText() []byte {
	data, _ := t.Time.MarshalText()
	return data
}

// ToTime convert string to time.Time
func ToTime(s string, layouts ...string) (t time.Time, err error) {
	value, layout := toTimeLayout(s, layouts...)
	if layout == "" {
		err = ErrConvertFail
		return
	}
	t, err = time.Parse(layout, value)
	return
}

// ToLocalTime convert string to time.Time in Local locale, add +8 hours.
func ToLocalTime(s string, layouts ...string) (t time.Time, err error) {
	t, err = ToTime(s, layouts...)
	if err == nil {
		return toLocalTime(t).Local(), nil
	}
	return
}

// ToUTCTime convert string to time.Time in UTC locale, add -8 hours.
func ToUTCTime(s string, layouts ...string) (t time.Time, err error) {
	t, err = ToTime(s, layouts...)
	if err == nil {
		return toUTCTime(t).UTC(), nil
	}
	return
}

func toTimeLayout(s string, layouts ...string) (value string, layout string) {
	value = s
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
		case 14:
			layout = TimeFormatStringS
		case 16:
			layout = DateTimeFormatStringM
		case 17:
			value = value[0:14] + "." + value[14:]
			layout = TimeFormatStringM
		case 18:
			layout = TimeFormatStringM
		case 19:
			layout = DateTimeFormatString
		case 20, 25:
			if strings.ContainsRune(s, '+') {
				layout = "2006-01-02 15:04:05-07:00"
			} else {
				layout = time.RFC3339
			}
		case 23:
			layout = TimeFormatString
		case 29, 34:
			layout = "2006-01-02 15:04:05.999999999-07:00"
		case 30, 35:
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
	return
}

// toLocalTime Convert time, add +8 hours.
func toLocalTime(t time.Time) time.Time {
	return t.Add(ZoneOffsetSecond)
}

// toUTCTime Convert time, add -8 hours.
func toUTCTime(t time.Time) time.Time {
	return t.Add(-1 * ZoneOffsetSecond)
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
