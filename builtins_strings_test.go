package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_builtin_string_functions(t *testing.T) {
	tests := []struct {
		expr   string
		result any
	}{
		{
			expr:   `substring ( "foobar" , 2 , 2)`,
			result: "oo",
		},
		{
			expr:   `string length("foobar")`,
			result: 6,
		},
		{
			expr:   `upper case ("foobar")`,
			result: "FOOBAR",
		},
		{
			expr:   `lower case ("FOOBAR")`,
			result: "foobar",
		},
		{
			expr:   `substring before("foobar", "b")`,
			result: "foo",
		},
		{
			expr:   `substring after("foobar", "b")`,
			result: "ar",
		},
		{
			expr:   `replace ( "fooXXbar" , "XX" , "" )`,
			result: "foobar",
		},
		{
			expr:   `contains ("foobar", "oo")`,
			result: true,
		},
		{
			expr:   `starts with ("foobar", "foo")`,
			result: true,
		},
		{
			expr:   `ends with ("foobar", "bar")`,
			result: true,
		},
		{
			expr:   `matches("foobar", "^foo")`,
			result: true,
		},
		{
			expr:   `split("foo,bar", ",")`,
			result: []string{"foo", "bar"},
		},
		{
			expr:   `string join(["foo","bar"], "-")`,
			result: "foo-bar",
		},
		{
			expr:   `string join(["foo","bar"])`,
			result: "foobar",
		},
		{
			expr:   `string(123)`,
			result: "123",
		},
		{
			expr:   `"foo" + "bar"`,
			result: "foobar",
		},
	}
	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			actual, err := EvalString(test.expr)
			assert.Nil(t, err)
			assert.Equal(t, test.result, actual)
		})
	}
}
