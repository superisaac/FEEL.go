package feel

import (
	"fmt"
)

func installBuiltinFunctions(prelude *Prelude) {
	// conversion functions
	prelude.BindNativeFunc("string", func(v interface{}) (string, error) {
		return fmt.Sprintf("%s", v), nil
	}, []string{"from"})

	prelude.BindNativeFunc("number", func(v interface{}) (*Number, error) {
		return ParseNumberWithErr(v)
	}, []string{"from"})

	// boolean functions
	prelude.BindNativeFunc("not", func(v interface{}) (bool, error) {
		return !boolValue(v), nil
	}, []string{"from"})

	prelude.BindMacro("is defined", func(intp *Interpreter, args []AST) (interface{}, error) {
		if varNode, ok := args[0].(*Var); ok {
			if _, ok := intp.Resolve(varNode.Name); !ok {
				return false, nil
			}
		}
		// TODO: more condition tests
		return true, nil
	}, []string{"value"})
}
