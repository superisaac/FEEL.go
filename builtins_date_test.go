package feel

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_builtin_date_functions_is(t *testing.T) {
	actual, err := EvalString(`is( date( "2012-12-25" ), date( "2012-12-25" ) )`)
	assert.Nil(t, err)
	assert.Equal(t, true, actual)
}
