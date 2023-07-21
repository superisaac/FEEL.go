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

func (self FEELTime) Time() time.Time {
	return self.t
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

func (self FEELTime) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "hour":
		return self.t.Hour(), true
	case "minute":
		return self.t.Minute(), true
	case "second":
		return self.t.Second(), true
	case "timezone":
		zoneName, _ := self.t.Zone()
		return zoneName, true
	case "timezone offset":
		_, offset := self.t.Zone()
		return offset, true
	}
	return nil, false
}

func (self FEELTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELTime) String() string {
	return self.t.Format("15:04:05-07:00")
}

// Date
type FEELDate struct {
	t time.Time
}

func (self FEELDate) Date() time.Time {
	return self.t
}

func (self FEELDate) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "year":
		return self.t.Year(), true
	case "month":
		return self.t.Month(), true
	case "day":
		return self.t.Day(), true
	}
	return nil, false
}

func (self FEELDate) String() string {
	return self.t.Format("2006-01-02")
}

func (self FEELDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
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

func (self FEELDatetime) Time() time.Time {
	return self.t
}

func (self FEELDatetime) Date() time.Time {
	return self.t
}

func (self FEELDatetime) Equal(other FEELDatetime) bool {
	return self.t.Equal(other.t)
}

func (self FEELDatetime) Compare(other FEELDatetime) int {
	if self.t.Equal(other.t) {
		return 0
	} else if self.t.Before(other.t) {
		return -1
	} else {
		return 1
	}
}

func (self FEELDatetime) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "year":
		return self.t.Year(), true
	case "month":
		return self.t.Month(), true
	case "day":
		return self.t.Day(), true
	case "hour":
		return self.t.Hour(), true
	case "minute":
		return self.t.Minute(), true
	case "second":
		return self.t.Second(), true
	case "timezone":
		zoneName, _ := self.t.Zone()
		return zoneName, true
	case "timezone offset":
		_, offset := self.t.Zone()
		return offset, true
	}
	return nil, false
}

func (self FEELDatetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELDatetime) String() string {
	return self.t.Format("2006-01-02T15:04:05@MST")
}

func (self *FEELDatetime) Add(dur *FEELDuration) *FEELDatetime {
	return &FEELDatetime{t: self.t.Add(dur.Duration())}
}

func (self *FEELDatetime) Sub(v HasTime) *FEELDuration {
	delta := self.t.Sub(v.Time())
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

func (self FEELDuration) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "years":
		return self.Years, true
	case "months":
		return self.Months, true
	case "days":
		return self.Days, true
	case "hours":
		return self.Hours, true
	case "minutes":
		return self.Minutes, true
	case "seconds":
		return self.Seconds, true
	}
	return nil, false
}

func (self FEELDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELDuration) Duration() time.Duration {
	// self.Year and self.Month
	dv := (time.Duration(self.Days*24+self.Hours)*time.Hour +
		time.Duration(self.Minutes)*time.Minute +
		time.Duration(self.Seconds)*time.Second)
	if self.Neg {
		dv = -dv
	}
	return dv
}

func (self *FEELDuration) Negative() *FEELDuration {
	neg := *self
	neg.Neg = !self.Neg
	return &neg
}

func (self FEELDuration) String() string {
	sYear, sMonth, sDay, sHour, sMinute, sSecond := "", "", "", "", "", ""
	sNeg := ""
	if self.Neg {
		sNeg = "-"
	}
	if self.Years != 0 {
		sYear = fmt.Sprintf("%dY", self.Years)
	}
	if self.Months != 0 {
		sMonth = fmt.Sprintf("%dM", self.Months)
	}
	if self.Days != 0 {
		sDay = fmt.Sprintf("%dD", self.Days)
	}

	if self.Hours != 0 {
		sHour = fmt.Sprintf("%dH", self.Hours)
	}
	if self.Minutes != 0 {
		sMinute = fmt.Sprintf("%dM", self.Minutes)
	}
	if self.Seconds != 0 {
		sSecond = fmt.Sprintf("%dS", self.Seconds)
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
	prelude.BindNativeFunc("date", func(intp *Interpreter, frm string) (interface{}, error) {
		return ParseDate(frm)
	}, []string{"from"})

	prelude.BindNativeFunc("time", func(intp *Interpreter, frm string) (interface{}, error) {
		return ParseTime(frm)
	}, []string{"from"})

	prelude.BindNativeFunc("date and time", func(intp *Interpreter, frm string) (interface{}, error) {
		return ParseDatetime(frm)
	}, []string{"from"})

	prelude.BindNativeFunc("duration", func(intp *Interpreter, frm string) (interface{}, error) {
		return ParseDuration(frm)
	}, []string{"from"})

	// temporal functions
	prelude.BindNativeFunc("now", func(intp *Interpreter) (interface{}, error) {
		return &FEELDatetime{t: time.Now()}, nil
	}, nil)

	prelude.BindNativeFunc("today", func(intp *Interpreter) (interface{}, error) {
		return &FEELDate{t: time.Now()}, nil
	}, nil)

	prelude.BindNativeFunc("day of week", func(intp *Interpreter, v HasDate) (interface{}, error) {
		return v.Date().Weekday(), nil
	}, []string{"date"})

	prelude.BindNativeFunc("day of year", func(intp *Interpreter, v HasDate) (interface{}, error) {
		return v.Date().YearDay(), nil
	}, []string{"date"})

	prelude.BindNativeFunc("week of year", func(intp *Interpreter, v HasDate) (interface{}, error) {
		_, week := v.Date().ISOWeek()
		return week, nil
	}, []string{"date"})

	prelude.BindNativeFunc("month of year", func(intp *Interpreter, v HasDate) (interface{}, error) {
		return v.Date().Month(), nil
	}, []string{"date"})

	prelude.BindNativeFunc("abs", func(intp *Interpreter, dur *FEELDuration) (interface{}, error) {
		newDur := *dur
		newDur.Neg = false
		return newDur, nil
	}, []string{"dur"})

	// refs https://docs.camunda.io/docs/components/modeler/feel/builtin-functions/feel-built-in-functions-temporal/#last-day-of-monthdate
	prelude.BindNativeFunc("last day of month", func(intp *Interpreter, v HasDate) (interface{}, error) {
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
	}, []string{"date"})
}
