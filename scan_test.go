package feel

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleScannerFinding(t *testing.T) {
	assert := assert.New(t)

	input := `
	< alice + bob 
	// single line comment
	/** multi
	line
	comment
	 */

	+ "op\nq" * true ( haha )

+ "multiline
	string中文"

	中文变量andγβ
	`

	scanner := NewScanner(input)
	tokens, err := scanner.Tokens()
	assert.Nil(err)
	// tokens is ["<", "name", "+", "name"]
	assert.Equal("<", tokens[0].Kind)

	assert.Equal("name", tokens[1].Kind)
	assert.Equal("alice", tokens[1].Value)

	assert.Equal("+", tokens[2].Kind)

	assert.Equal("name", tokens[3].Kind)
	assert.Equal("bob", tokens[3].Value)

	assert.Equal("+", tokens[4].Kind)

	assert.Equal("string", tokens[5].Kind)
	assert.Equal(`"op\nq"`, tokens[5].Value)
	assert.Equal(7, len(tokens[5].Value))

	assert.Equal("*", tokens[6].Kind)

	assert.Equal("keyword", tokens[7].Kind)
	assert.Equal("true", tokens[7].Value)

	assert.Equal("(", tokens[8].Kind)

	assert.Equal("name", tokens[9].Kind)
	assert.Equal("haha", tokens[9].Value)

	assert.Equal(")", tokens[10].Kind)

	assert.Equal("string", tokens[12].Kind)
	assert.Equal("\"multiline\n\tstring中文\"", tokens[12].Value)

	assert.Equal("name", tokens[13].Kind)
	assert.Equal("中文变量andγβ", tokens[13].Value)

	// test against position
	assert.Equal(14, scanner.Pos.Row)
	assert.Equal(1, scanner.Pos.Column)
}

func TestUnicodeRegexp(t *testing.T) {
	assert := assert.New(t)

	reName := regexp.MustCompile(`\p{Han}+`)

	found := reName.FindString("abc(汉字)")
	assert.Equal("汉字", found)
}
