package feel

// refer to https://kiegroup.github.io/dmn-feel-handbook/#date
// refer to https://docs.camunda.io/docs/components/modeler/feel/language-guide/feel-temporal-expressions/

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var ErrParseTemporal = errors.New("fail to parse temporal value")

type HasTime interface {
	Time() time.Time
}

type HasDate interface {
	Date() time.Time
}

// time
type FEELTime struct {
	t time.Time
}

func (st FEELTime) Time() time.Time {
	return st.t
}

var timePatterns = []string{
	"15:04:05",
	"15:04:05-07:00",
	"15:04:05@MST",
}

func ParseTime(temporalStr string) (*FEELTime, error) {
	for _, pat := range timePatterns {
		if t, err := time.Parse(pat, temporalStr); err == nil {
			return &FEELTime{t: t}, nil
		}
	}
	return nil, ErrParseTemporal
}

func (st FEELTime) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "hour":
		return st.t.Hour(), true
	case "minute":
		return st.t.Minute(), true
	case "second":
		return st.t.Second(), true
	case "timezone":
		zoneName, _ := st.t.Zone()
		return zoneName, true
	case "timezone offset":
		_, offset := st.t.Zone()
		return offset, true
	}
	return nil, false
}

func (st FEELTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(st.String())
}

func (st FEELTime) String() string {
	return st.t.Format("15:04:05-07:00")
}

// Date
type FEELDate struct {
	t time.Time
}

func (date FEELDate) Date() time.Time {
	return date.t
}

func (date FEELDate) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "year":
		return date.t.Year(), true
	case "month":
		return date.t.Month(), true
	case "day":
		return date.t.Day(), true
	}
	return nil, false
}

func (date FEELDate) String() string {
	return date.t.Format("2006-01-02")
}

func (date FEELDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(date.String())
}

var datePatterns = []string{
	"2006-01-02",
}

func ParseDate(timeStr string) (*FEELDate, error) {
	for _, pat := range datePatterns {
		if t, err := time.Parse(pat, timeStr); err == nil {
			return &FEELDate{t: t}, nil
		}
	}
	return nil, ErrParseTemporal
}

// Date and Time
type FEELDatetime struct {
	t time.Time
}

func (sdt FEELDatetime) Time() time.Time {
	return sdt.t
}

func (sdt FEELDatetime) Date() time.Time {
	return sdt.t
}

func (sdt FEELDatetime) Equal(other FEELDatetime) bool {
	return sdt.t.Equal(other.t)
}

func (sdt FEELDatetime) Compare(other FEELDatetime) int {
	if sdt.t.Equal(other.t) {
		return 0
	} else if sdt.t.Before(other.t) {
		return -1
	} else {
		return 1
	}
}

func (sdt FEELDatetime) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "year":
		return sdt.t.Year(), true
	case "month":
		return sdt.t.Month(), true
	case "day":
		return sdt.t.Day(), true
	case "hour":
		return sdt.t.Hour(), true
	case "minute":
		return sdt.t.Minute(), true
	case "second":
		return sdt.t.Second(), true
	case "timezone":
		zoneName, _ := sdt.t.Zone()
		return zoneName, true
	case "timezone offset":
		_, offset := sdt.t.Zone()
		return offset, true
	}
	return nil, false
}

func (sdt FEELDatetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(sdt.String())
}

func (sdt FEELDatetime) String() string {
	return sdt.t.Format("2006-01-02T15:04:05@MST")
}

func (sdt *FEELDatetime) Add(dur *FEELDuration) *FEELDatetime {
	if dur.Years > 0 || dur.Months > 0 {
		durMonths := dur.Years*12 + dur.Months
		timeMonths := sdt.t.Year()*12 + int(sdt.t.Month()-1)

		newTimeMonths := timeMonths + durMonths
		if dur.Neg {
			newTimeMonths = timeMonths - durMonths
		}
		return &FEELDatetime{
			t: time.Date(
				newTimeMonths/12, time.Month(newTimeMonths%12+1),
				sdt.t.Day(), sdt.t.Hour(), sdt.t.Minute(),
				sdt.t.Second(), sdt.t.Nanosecond(),
				sdt.t.Location()),
		}
	}
	return &FEELDatetime{t: sdt.t.Add(dur.Duration())}
}

func (sdt *FEELDatetime) Sub(v HasTime) *FEELDuration {
	delta := sdt.t.Sub(v.Time())
	return NewFEELDuration(delta)
}

var dateTimePatterns = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02T15:04:05@MST",
}

func ParseDatetime(temporalStr string) (*FEELDatetime, error) {
	for _, pat := range dateTimePatterns {
		if t, err := time.Parse(pat, temporalStr); err == nil {
			return &FEELDatetime{t: t}, nil
		}
	}
	return nil, ErrParseTemporal
}

func MustParseDatetime(temporalStr string) *FEELDatetime {
	t, err := ParseDatetime(temporalStr)
	if err != nil {
		panic(err)
	}
	return t
}

type FEELDuration struct {
	Neg     bool
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

func NewFEELDuration(dur time.Duration) *FEELDuration {
	d := &FEELDuration{}
	ndur := int(dur)
	nhours := ndur / int(time.Hour)
	remain := ndur - nhours*int(time.Hour)
	nmins := remain / int(time.Minute)

	remain -= nmins * int(time.Minute)
	nsecs := remain / int(time.Second)

	d.Days = nhours / 24
	d.Hours = nhours - d.Days*24
	d.Minutes = nmins
	d.Seconds = nsecs
	return d
}

func (dur FEELDuration) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "years":
		return dur.Years, true
	case "months":
		return dur.Months, true
	case "days":
		return dur.Days, true
	case "hours":
		return dur.Hours, true
	case "minutes":
		return dur.Minutes, true
	case "seconds":
		return dur.Seconds, true
	}
	return nil, false
}

func (dur FEELDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(dur.String())
}

func (dur FEELDuration) Duration() time.Duration {
	// dur.Year and dur.Month
	dv := (time.Duration(dur.Days*24+dur.Hours)*time.Hour +
		time.Duration(dur.Minutes)*time.Minute +
		time.Duration(dur.Seconds)*time.Second)
	if dur.Neg {
		dv = -dv
	}
	return dv
}

func (dur *FEELDuration) Negative() *FEELDuration {
	neg := *dur
	neg.Neg = !dur.Neg
	return &neg
}

func (dur FEELDuration) String() string {
	sYear, sMonth, sDay, sHour, sMinute, sSecond := "", "", "", "", "", ""
	sNeg := ""
	if dur.Neg {
		sNeg = "-"
	}
	if dur.Years != 0 {
		sYear = fmt.Sprintf("%dY", dur.Years)
	}
	if dur.Months != 0 {
		sMonth = fmt.Sprintf("%dM", dur.Months)
	}
	if dur.Days != 0 {
		sDay = fmt.Sprintf("%dD", dur.Days)
	}

	if dur.Hours != 0 {
		sHour = fmt.Sprintf("%dH", dur.Hours)
	}
	if dur.Minutes != 0 {
		sMinute = fmt.Sprintf("%dM", dur.Minutes)
	}
	if dur.Seconds != 0 {
		sSecond = fmt.Sprintf("%dS", dur.Seconds)
	}
	if sYear != "" || sMonth != "" {
		return fmt.Sprintf("%sP%s%s", sNeg, sYear, sMonth)
	} else {
		return fmt.Sprintf("%sP%sT%s%s%s", sNeg, sDay, sHour, sMinute, sSecond)
	}
}

var yearmonthDurationPattern = regexp.MustCompile(`^(\-?)P((\d+)Y)?((\d+)M)?$`)
var timeDurationPattern = regexp.MustCompile(`^(\-?)P((\d+)D)?T((\d+)H)?((\d+)M)?((\d+)S)?$`)

func ParseDuration(temporalStr string) (*FEELDuration, error) {
	// parse year month duration
	if submatches := yearmonthDurationPattern.FindStringSubmatch(temporalStr); submatches != nil {
		dur := &FEELDuration{}
		if submatches[1] != "" {
			dur.Neg = true
		}

		if submatches[2] != "" {
			y, err := strconv.ParseInt(submatches[3], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Years = int(y)
		}
		if submatches[4] != "" {
			m, err := strconv.ParseInt(submatches[5], 10, 64)
			if err != nil {
				return nil, err
			}
			if m > 12 {
				return nil, errors.New("months cannot exceed 12")
			}
			dur.Months = int(m)
		}
		return dur, nil
	}

	// parse day time duration
	if submatches := timeDurationPattern.FindStringSubmatch(temporalStr); submatches != nil {
		dur := &FEELDuration{}
		if submatches[1] != "" {
			dur.Neg = true
		}
		if submatches[2] != "" {
			v, err := strconv.ParseInt(submatches[3], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Days = int(v)
		}
		if submatches[4] != "" {
			v, err := strconv.ParseInt(submatches[5], 10, 64)
			if err != nil {
				return nil, err
			}
			if v > 24 {
				return nil, errors.New("hours cannot exceed 24")
			}
			dur.Hours = int(v)
		}
		if submatches[6] != "" {
			v, err := strconv.ParseInt(submatches[7], 10, 64)
			if err != nil {
				return nil, err
			}
			if v > 60 {
				return nil, errors.New("minutes cannot exceed 60")
			}
			dur.Minutes = int(v)
		}
		if submatches[8] != "" {
			v, err := strconv.ParseInt(submatches[9], 10, 64)
			if err != nil {
				return nil, err
			}
			if v > 60 {
				return nil, errors.New("seconds cannot exceed 60")
			}
			dur.Seconds = int(v)
		}
		return dur, nil
	}

	return nil, ErrParseTemporal
}

func ParseTemporalValue(temporalStr string) (interface{}, error) {
	if v, err := ParseDatetime(temporalStr); err == nil {
		return v, nil
	}

	if v, err := ParseTime(temporalStr); err == nil {
		return v, nil
	}

	if v, err := ParseDate(temporalStr); err == nil {
		return v, nil
	}

	return ParseDuration(temporalStr)
}

// builtin functions
func installDatetimeFunctions(prelude *Prelude) {
	// conversions
	prelude.Bind("date", wrapTyped(func(frm string) (interface{}, error) {
		return ParseDate(frm)
	}).Required("from"))

	prelude.Bind("time", wrapTyped(func(frm string) (interface{}, error) {
		return ParseTime(frm)
	}).Required("from"))

	prelude.Bind("date and time", wrapTyped(func(frm string) (interface{}, error) {
		return ParseDatetime(frm)
	}).Required("from"))

	prelude.Bind("duration", wrapTyped(func(frm string) (interface{}, error) {
		return ParseDuration(frm)
	}).Required("from"))

	// temporal functions
	prelude.Bind("now", wrapTyped(func() (interface{}, error) {
		return &FEELDatetime{t: time.Now()}, nil
	}))

	prelude.Bind("today", wrapTyped(func() (interface{}, error) {
		return &FEELDate{t: time.Now()}, nil
	}))

	prelude.Bind("day of week", wrapTyped(func(v HasDate) (interface{}, error) {
		return v.Date().Weekday(), nil
	}).Required("date"))

	prelude.Bind("day of year", wrapTyped(func(v HasDate) (interface{}, error) {
		return v.Date().YearDay(), nil
	}).Required("date"))

	prelude.Bind("week of year", wrapTyped(func(v HasDate) (interface{}, error) {
		_, week := v.Date().ISOWeek()
		return week, nil
	}).Required("date"))

	prelude.Bind("month of year", wrapTyped(func(v HasDate) (interface{}, error) {
		return v.Date().Month(), nil
	}).Required("date"))

	prelude.Bind("abs", wrapTyped(func(dur *FEELDuration) (interface{}, error) {
		newDur := *dur
		newDur.Neg = false
		return newDur, nil
	}).Required("dur"))

	// refs https://docs.camunda.io/docs/components/modeler/feel/builtin-functions/feel-built-in-functions-temporal/#last-day-of-monthdate
	prelude.Bind("last day of month", wrapTyped(func(v HasDate) (interface{}, error) {
		month := v.Date().Month()
		year := v.Date().Year()
		if month == 12 {
			year++
			month = 1
		} else {
			month++
		}
		nextFirstDay := time.Date(year, month, 1, 0, 0, 0, 0, v.Date().Location())
		lastDay := nextFirstDay.Add(-24 * time.Hour) // 1 day before
		return lastDay.Day(), nil
	}).Required("date"))
}
