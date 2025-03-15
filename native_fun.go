package feel

import (
	"fmt"
	"sync"
)

// arg type error
type ArgTypeError struct {
	index  int
	expect string
}

func (err ArgTypeError) Error() string {
	return fmt.Sprintf("type error at %d, expect %s", err.index, err.expect)
}

// ArgSizeError
type ArgSizeError struct {
	has    int
	expect int
}

func (err ArgSizeError) Error() string {
	return fmt.Sprintf("argument size error, has %d, expect %d", err.has, err.expect)
}

// native function
type NativeFunDef func(args map[string]interface{}) (interface{}, error)

type NativeFun struct {
	fn               NativeFunDef
	requiredArgNames []string
	optionalArgNames []string
	varArgName       string
	help             string
}

func NewNativeFunc(fn NativeFunDef) *NativeFun {
	return &NativeFun{fn: fn}
}

func (nfun *NativeFun) Required(argNames ...string) *NativeFun {
	nfun.requiredArgNames = append(nfun.requiredArgNames, argNames...)
	return nfun
}

func (nfun *NativeFun) Optional(argNames ...string) *NativeFun {
	nfun.optionalArgNames = append(nfun.optionalArgNames, argNames...)
	return nfun
}

func (nfun *NativeFun) Vararg(argName string) *NativeFun {
	nfun.varArgName = argName
	return nfun
}

func (nfun *NativeFun) Help(help string) *NativeFun {
	nfun.help = help
	return nfun
}

func (nfun NativeFun) ArgNameAt(at int) (string, bool) {
	if at >= 0 && at < len(nfun.requiredArgNames) {
		return nfun.requiredArgNames[at], true
	} else if at >= len(nfun.requiredArgNames) && at < len(nfun.requiredArgNames)+len(nfun.optionalArgNames) {
		return nfun.optionalArgNames[at-len(nfun.requiredArgNames)], true
	}
	return "", false
}

func (nfun *NativeFun) Call(intp *Interpreter, args map[string]interface{}) (interface{}, error) {
	v, err := nfun.fn(args)
	if err != nil {
		return nil, err
	}
	return normalizeValue(v), nil
}

// macro
type MacroDef func(intp *Interpreter, args map[string]Node, varArgs []Node) (interface{}, error)
type Macro struct {
	fn               MacroDef
	requiredArgNames []string
	optionalArgNames []string
	varArgName       string
	help             string
}

func NewMacro(fn MacroDef) *Macro {
	return &Macro{fn: fn}
}

func (macro *Macro) Required(argNames ...string) *Macro {
	macro.requiredArgNames = append(macro.requiredArgNames, argNames...)
	return macro
}

func (macro *Macro) Optional(argNames ...string) *Macro {
	macro.optionalArgNames = append(macro.optionalArgNames, argNames...)
	return macro
}
func (macro *Macro) Vararg(argName string) *Macro {
	macro.varArgName = argName
	return macro
}
func (macro *Macro) Help(help string) *Macro {
	macro.help = help
	return macro
}

// Prelude
type Prelude struct {
	vars map[string]interface{}
}

var loadOnce sync.Once
var inst *Prelude

func GetPrelude() *Prelude {
	loadOnce.Do(func() {
		inst = &Prelude{vars: make(map[string]interface{})}
		inst.Load()
	})
	return inst
}

func (prelude *Prelude) Load() {
	// prelude.Bind("bind", NewMacro(func(intp *Interpreter, args map[string]Node, varArgs []Node) (any, error) {
	// 	name, _ := args["name"].Eval(intp)
	// 	strName, ok := name.(string)
	// 	if !ok {
	// 		return nil, NewErrTypeMismatch("string")
	// 	}
	// 	v, err := args["value"].Eval(intp)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	intp.Bind(strName, v)
	// 	return v, nil
	// }).Required("name", "value").Help("bind value to name in current top scope"))

	// prelude.Bind("set", NewMacro(func(intp *Interpreter, args map[string]Node, varArgs []Node) (any, error) {
	// 	name, _ := args["name"].Eval(intp)
	// 	strName, ok := name.(string)
	// 	if !ok {
	// 		return nil, NewErrTypeMismatch("string")
	// 	}
	// 	v, err := args["value"].Eval(intp)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if intp.Set(strName, v) {
	// 		return v, nil
	// 	} else {
	// 		intp.Bind(strName, v)
	// 		return v, nil
	// 	}
	// }).Required("name", "value").Help("bind value to name in resolved scope, if not found, it's bind to current top scope(the same as 'bind')"))

	prelude.Bind("block", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
		var lastValue interface{}
		var err error
		for _, expr := range exprlist {
			lastValue, err = expr.Eval(intp)
			if err != nil {
				return nil, err
			}
		}
		return lastValue, nil
	}).Vararg("express list").Help("quote a sequence of expresses and return the last result"))

	prelude.Bind("help", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		switch vv := v.(type) {
		case *NativeFun:
			return vv.help, nil
		case *Macro:
			return vv.help, nil
		default:
			return typeName(vv), nil
		}
	}).Required("value").Help("the help information of a value"))

	prelude.Bind("typeof", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		return typeName(v), nil
	}).Required("value").Help("the type of a value"))

	installDatetimeFunctions(prelude)
	installBuiltinFunctions(prelude)
	installContextFunctions(prelude)
	installRangeFunctions(prelude)
}

func (prelude *Prelude) Bind(name string, value interface{}) *Prelude {
	if _, ok := prelude.vars[name]; ok {
		panic(fmt.Sprintf("bind(), name '%s' already bound", name))
	}
	prelude.vars[name] = normalizeValue(value)
	return prelude
}

func (prelude *Prelude) Resolve(name string) (interface{}, bool) {
	v, ok := prelude.vars[name]
	return v, ok
}

// buildin native funcs
func nativeBind(intp *Interpreter, varname string, value interface{}) (interface{}, error) {
	intp.Bind(varname, value)
	return nil, nil
}
