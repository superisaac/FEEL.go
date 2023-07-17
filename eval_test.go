package feel

import (
	"fmt"
	"gotest.tools/assert"
	"testing"
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

	dt, err := ParseDateTime(`2023-06-07T15:04:05`)
	assert.NilError(t, err)
	assert.DeepEqual(t, dt.t.Hour(), 15)
	assert.DeepEqual(t, dt.t.Second(), 5)

	dur, err := ParseDuration("P12Y2M")
	assert.NilError(t, err)
	assert.Equal(t, 12, dur.Year)
	assert.Equal(t, 2, dur.Month)

	dur1, err := ParseDuration("P7M")
	assert.NilError(t, err)
	assert.Equal(t, 0, dur1.Year)
	assert.Equal(t, 7, dur1.Month)
}
