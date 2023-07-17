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

// time
type FEELTime struct {
	t time.Time
}

func (self FEELTime) Time() time.Time {
	return self.t
}

var timePatterns = []string{
	"15:04:05",
	"15:04:05+07:00",
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

func (self FEELTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELTime) String() string {
	return self.t.Format("15:04:05+07:00")
}

// Date
type FEELDate struct {
	t time.Time
}

func (self FEELDate) Time() time.Time {
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
type FEELDateTime struct {
	t time.Time
}

func (self FEELDateTime) Time() time.Time {
	return self.t
}

func (self FEELDateTime) GetAttr(name string) (interface{}, bool) {
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

func (self FEELDateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELDateTime) String() string {
	return self.t.Format("2006-01-02T15:04:05@MST")
}

var dateTimePatterns = []string{
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05+07:00",
	"2006-01-02T15:04:05@MST",
}

func ParseDateTime(temporalStr string) (*FEELDateTime, error) {
	for _, pat := range dateTimePatterns {
		if t, err := time.Parse(pat, temporalStr); err == nil {
			return &FEELDateTime{t: t}, nil
		}
	}
	return nil, ErrParseTemporal
}

type FEELDuration struct {
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Second int
}

func (self FEELDuration) GetAttr(name string) (interface{}, bool) {
	switch name {
	case "year":
		return self.Year, true
	case "month":
		return self.Month, true
	case "day":
		return self.Day, true
	case "hour":
		return self.Hour, true
	case "minute":
		return self.Minute, true
	case "second":
		return self.Second, true
	}
	return nil, false
}

func (self FEELDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.String())
}

func (self FEELDuration) String() string {
	sYear, sMonth, sDay, sHour, sMinute, sSecond := "", "", "", "", "", ""

	if self.Year != 0 {
		sYear = fmt.Sprintf("%dY", self.Year)
	}
	if self.Month != 0 {
		sMonth = fmt.Sprintf("%dM", self.Month)
	}
	if self.Day != 0 {
		sDay = fmt.Sprintf("%dD", self.Day)
	}

	if self.Hour != 0 {
		sDay = fmt.Sprintf("%dH", self.Hour)
	}
	if self.Minute != 0 {
		sDay = fmt.Sprintf("%dM", self.Minute)
	}
	if self.Second != 0 {
		sDay = fmt.Sprintf("%dS", self.Second)
	}
	if sYear != "" || sMonth != "" {
		return fmt.Sprintf("P%s%s", sYear, sMonth)
	} else {
		return fmt.Sprintf("P%sT%s%s%s", sDay, sHour, sMinute, sSecond)
	}

}

var yearmonthDurationPattern = regexp.MustCompile(`^P((\d+)Y)?((\d+)M)?$`)
var timeDurationPatteren = regexp.MustCompile(`^P((\d+)D)?T((\d+)H)?((\d+)M)?((\d+)S)?$`)

func ParseDuration(temporalStr string) (*FEELDuration, error) {
	// parse year month duration
	if submatches := yearmonthDurationPattern.FindStringSubmatch(temporalStr); submatches != nil {
		dur := &FEELDuration{}
		if submatches[1] != "" {
			y, err := strconv.ParseInt(submatches[2], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Year = int(y)
		}
		if submatches[3] != "" {
			m, err := strconv.ParseInt(submatches[4], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Month = int(m)
		}
		return dur, nil
	}

	// parse day time duration
	if submatches := timeDurationPatteren.FindStringSubmatch(temporalStr); submatches != nil {
		dur := &FEELDuration{}
		if submatches[1] != "" {
			v, err := strconv.ParseInt(submatches[2], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Day = int(v)
		}
		if submatches[3] != "" {
			v, err := strconv.ParseInt(submatches[4], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Hour = int(v)
		}
		if submatches[5] != "" {
			v, err := strconv.ParseInt(submatches[6], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Minute = int(v)
		}
		if submatches[7] != "" {
			v, err := strconv.ParseInt(submatches[8], 10, 64)
			if err != nil {
				return nil, err
			}
			dur.Second = int(v)
		}
		return dur, nil
	}

	return nil, ErrParseTemporal
}

func ParseTemporalValue(temporalStr string) (interface{}, error) {
	if v, err := ParseDateTime(temporalStr); err == nil {
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
func installDateTimeFunctions(prelude *Prelude) {
	prelude.BindNativeFunc("now", func(intp *Interpreter) (interface{}, error) {
		return &FEELDateTime{t: time.Now()}, nil
	})

	prelude.BindNativeFunc("today", func(intp *Interpreter) (interface{}, error) {
		return &FEELDate{t: time.Now()}, nil
	})

	prelude.BindNativeFunc("day of week", func(intp *Interpreter, v HasTime) (interface{}, error) {
		return v.Time().Weekday(), nil
	}, "date")

	prelude.BindNativeFunc("day of year", func(intp *Interpreter, v HasTime) (interface{}, error) {
		return v.Time().YearDay(), nil
	}, "date")

	prelude.BindNativeFunc("week of year", func(intp *Interpreter, v HasTime) (interface{}, error) {
		_, week := v.Time().ISOWeek()
		return week, nil
	}, "date")

	prelude.BindNativeFunc("month of year", func(intp *Interpreter, v HasTime) (interface{}, error) {
		return v.Time().Month(), nil
	}, "date")

	// TODO: abs, last day of month
}
