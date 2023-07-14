package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleParser(t *testing.T) {
	assert := assert.New(t)

	input := `
	abc + 3 * sum(2, 7) - @"2023-06-01 05:01:00@Asia/Shanghai"
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal(`(- (+ abc (* 3 (call sum [2, 7]))) @"2023-06-01 05:01:00@Asia/Shanghai")`, ast.Repr())

}

func TestStringVal(t *testing.T) {
	assert := assert.New(t)

	input := `
	"a string\n value"
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	strVal, ok := ast.(*StringNode)
	assert.True(ok)

	assert.Equal("a string\u000A value", strVal.Content())

}

func TestContBinop(t *testing.T) {
	assert := assert.New(t)

	input := `
	abc + 3 * u - eight.value
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(- (+ abc (* 3 u)) (. eight value))", ast.Repr())
}

func TestCompare(t *testing.T) {
	assert := assert.New(t)

	input := `
	a = 3
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(= a 3)", ast.Repr())
}
func TestIfExpression(t *testing.T) {
	assert := assert.New(t)

	input := `
	if a > 3 
	then "yes" else "no"
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(if (> a 3) \"yes\" \"no\")", ast.Repr())
}

func TestForExpression(t *testing.T) {
	assert := assert.New(t)

	input := `
	for x in [3, 4], y in [5, 9] return x * y
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(for x [3, 4] (for y [5, 9] (* x y)))", ast.Repr())
}

func TestSomeExpression(t *testing.T) {
	assert := assert.New(t)

	input := `
	some x in [3, 4, 5, 6, 9] satisfies x >= 5
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(some \"x\" [3, 4, 5, 6, 9] (>= x 5))", ast.Repr())
}

func TestRange(t *testing.T) {
	assert := assert.New(t)

	input := `
	(2..5]
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(2..5]", ast.Repr())

	input1 := `
	[2..5 * 9)
	`
	ast1, err := ParseString(input1)
	assert.Nil(err)
	assert.Equal("[2..(* 5 9))", ast1.Repr())
}

func TestUnaryTests(t *testing.T) {
	assert := assert.New(t)

	input := `
	(2..5], >= 6, 100, !=888
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal("(multitests (2..5] (>= ? 6) 100 (!= ? 888))", ast.Repr())
}

func TestFunCallAndIndex(t *testing.T) {
	assert := assert.New(t)

	input := `
	a.b.c(3, 4).d.e[4].f + 1001
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal(`(+ (. ([] (. (. (call (. (. a b) c) [3, 4]) d) e) 4) f) 1001)`, ast.Repr())

	input1 := `
	string contains("abc def" , "e")
	`
	ast1, err := ParseString(input1)
	assert.Nil(err)
	assert.Equal("(call `string contains` [\"abc def\", \"e\"])", ast1.Repr())
}

func TestFunDef(t *testing.T) {
	assert := assert.New(t)

	input := `
	function(a, b) a + b * 2
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal(`(function [a, b] (+ a (* b 2)))`, ast.Repr())

	_, err1 := ParseString(`function(a, b`)
	assert.NotNil(err1)
	une, ok := err1.(*UnexpectedToken)
	assert.True(ok)
	assert.Equal(TokenEOF, une.token.Kind)
	assert.Equal([]string{")", ","}, une.expects)
}

func TestMapValue(t *testing.T) {
	assert := assert.New(t)

	input := `
	{ a: 1, b: @"2023-06-01", c: [1, 2, 3]}
	`
	ast, err := ParseString(input)
	assert.Nil(err)
	assert.Equal(`(map ("a" 1) ("b" @"2023-06-01") ("c" [1, 2, 3]))`, ast.Repr())

	// parse incomplete inputs
	input1 := `
	{ a: 1, b: @"2023-06-01",
	`
	_, err1 := ParseString(input1)
	assert.NotNil(err1)
	une, ok := err1.(*UnexpectedToken)
	assert.True(ok)
	assert.Equal(TokenEOF, une.token.Kind)
	assert.Equal([]string{"name", "string"}, une.expects)
}
