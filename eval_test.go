package feel

import (
	"fmt"
	"testing"
	"time"

	"gotest.tools/assert"
)

type evalPair struct {
	input  string
	expect interface{}
}

func TestEvalPairs(t *testing.T) {
	//assert0 := assert.New(t)
	evalPairs := []evalPair{
		// empty input outputs nil
		{"", nil},

		{"5 + -6", ParseNumber(-1)},
		{"5 + 6", ParseNumber(11)},
		{"(function(a) 2 * a)(5)", ParseNumber(10)},
		{"true", true},
		{"false", false},
		{`"hello" + " world"`, "hello world"},

		{`{a if c: "hello", b: "world"}`, map[string]interface{}{"a if c": "hello", "b": "world"}},

		// in range and array
		{`5 in (5..8]`, false},
		{`5 in [5..8)`, true},
		{`8 in [5..8)`, false},
		{`8 in [5..8]`, true},

		{`"a" in ["a".."z"]`, true},
		{`5 in [3,5, 8]`, true},
		{`5 in [3, 6, 8]`, false},
		{`5 in []`, false},
		//{`not(5 in [3, 5, 9])`, false},

		// if then else
		{`bind("a", 5); if a > 3 then "larger" else "smaller"`, "larger"},
		{`bind("a", 5); if a = 5 then "equal" else "not equal"`, "equal"},
		{`bind("a b", 5); if a b = 5 then "equal" else "not equal"`, "equal"}, // a name has multiple chunks

		// test not
		{`not( 5 >  6)`, true},

		// loop functions
		{`some x in [3, 4, 5] satisfies x >= 4`, ParseNumber(4)},
		{`every y in [3, 4, 5] satisfies y >= 4`, []interface{}{ParseNumber(4), ParseNumber(5)}},

		// null check
		{`a != null and a.b > 10`, false},
		{`a = null or a.b > 10`, true},

		// keyword arguments
		{`bind("sub", function(a, b) a - b); sub(a: 4, b: 2)`, ParseNumber(2)},

		// temporal expressions
		{`last day of month(@"2020-02-11")`, ParseNumber(29)},
		{`last day of month(@"2021-01-07")`, ParseNumber(31)},
		{`last day of month(@"2023-06-11")`, ParseNumber(30)},
		{`last day of month(@"2023-07-11")`, ParseNumber(31)},

		{`@"2023-07-21T13:57:32@CST" - @"PT2H3M"`, MustParseDatetime("2023-07-21T11:54:32@CST")}, // test day/hour/min duration
		{`@"2023-06-01T10:33:20@CST" + @"P3Y11M"`, MustParseDatetime("2027-05-01T10:33:20@CST")}, // test year/month duration

		// builtin functions
		{`is defined(x)`, false},
		{`bind("x", 666); is defined(x)`, true},        // `x` is bound
		{`bind("x", 888); is defined(value: x)`, true}, // macro can use keyword arguments

		{`substring(string: "abcdef", start position: 2, length: 3)`, "cde"},
		{`substring(string: "abcdef", start position: 200, length: 3)`, ""},
		{`median([3, 5, 9, 1, "hello", -2])`, ParseNumber(3)},

		{`append(["hello"], " ", "world")`, []interface{}{"hello", " ", "world"}},
		{`concatenate([2, 1], [3])`, []interface{}{ParseNumber(2), ParseNumber(1), ParseNumber(3)}},
		{`insert before(["hello", "world"], 1, "another")`, []interface{}{"hello", "another", "world"}},
		{`remove(["hello", "a", "world"], 1)`, []interface{}{"hello", "world"}},

		{`sort(["hello", "a", "world"], function(x, y) x < y)`, []interface{}{"a", "hello", "world"}},
		{`sort([8, -1, 3], function(x, y) x > y)`, []interface{}{ParseNumber(8), ParseNumber(3), ParseNumber(-1)}},

		{`string join(["hello", "world"])`, "helloworld"},
		{`string join(["hello", "world"], " ", "[", "]")`, "[hello world]"},
	}

	for _, p := range evalPairs {
		res, err := EvalString(p.input)
		if err != nil {
			fmt.Printf("bad input %s\n", p.input)
		}
		assert.NilError(t, err)
		assert.DeepEqual(t, p.expect, res)
	}
}

func TestEvalUnaryTests(t *testing.T) {
	input := `> 8, <= 5`
	v, err := EvalStringWithScope(input, Scope{"?": 4})
	assert.NilError(t, err)
	assert.Equal(t, v, true)
}

func TestTemporalValue(t *testing.T) {
	input := `@"2023-06-07".day`
	v, err := EvalString(input)
	assert.NilError(t, err)
	assert.DeepEqual(t, v, ParseNumber(7))

	input1 := `@"2023-06-07T15:08:39".second`
	v1, err := EvalString(input1)
	assert.NilError(t, err)
	assert.DeepEqual(t, v1, ParseNumber(39))

	input2 := `@"P1DT3H25M60S".minutes`
	v2, err := EvalString(input2)
	assert.NilError(t, err)
	assert.DeepEqual(t, v2, ParseNumber(25))

	dt, err := ParseDatetime(`2023-06-07T15:04:05`)
	assert.NilError(t, err)
	assert.DeepEqual(t, dt.t.Hour(), 15)
	assert.DeepEqual(t, dt.t.Second(), 5)

	dur, err := ParseDuration("P12Y2M")
	assert.NilError(t, err)
	assert.Equal(t, 12, dur.Years)
	assert.Equal(t, 2, dur.Months)

	dur1, err := ParseDuration("P7M")
	assert.NilError(t, err)
	assert.Equal(t, 0, dur1.Years)
	assert.Equal(t, 7, dur1.Months)

	dur2, err := ParseDuration("PT20H")
	assert.NilError(t, err)
	assert.Equal(t, 20, dur2.Hours)
	assert.Equal(t, 0, dur2.Seconds)

	td, err := time.ParseDuration("3h37m20s")
	assert.NilError(t, err)
	dur3 := NewFEELDuration(td)
	assert.Equal(t, 3, dur3.Hours)
	assert.Equal(t, 37, dur3.Minutes)
	assert.Equal(t, 20, dur3.Seconds)
}
