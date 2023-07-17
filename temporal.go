package feel

// refer to https://kiegroup.github.io/dmn-feel-handbook/#date
// refer to https://docs.camunda.io/docs/components/modeler/feel/language-guide/feel-temporal-expressions/

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var ErrParseTemporal = errors.New("fail to parse temporal value")

// time
type FEELTime struct {
	t time.Time
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
	}
	return nil, false
}

// Date
type FEELDate struct {
	t time.Time
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
	}
	return nil, false
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
