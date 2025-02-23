package feel

import (
	"fmt"
	"sync"
)

type ArgTypeError struct {
	index  int
	expect string
}

func (self ArgTypeError) Error() string {
	return fmt.Sprintf("type error at %d, expect %s", self.index, self.expect)
}

type ArgSizeError struct {
	has    int
	expect int
}

func (self ArgSizeError) Error() string {
	return fmt.Sprintf("argument size error, has %d, expect %d", self.has, self.expect)
}

// NativeFunDef native function
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

func (nativeFun *NativeFun) Required(argNames ...string) *NativeFun {
	nativeFun.requiredArgNames = append(nativeFun.requiredArgNames, argNames...)
	return nativeFun
}

func (nativeFun *NativeFun) Optional(argNames ...string) *NativeFun {
	nativeFun.optionalArgNames = append(nativeFun.optionalArgNames, argNames...)
	return nativeFun
}

func (nativeFun *NativeFun) Vararg(argName string) *NativeFun {
	nativeFun.varArgName = argName
	return nativeFun
}

func (nativeFun *NativeFun) Help(help string) *NativeFun {
	nativeFun.help = help
	return nativeFun
}

func (nativeFun NativeFun) ArgNameAt(at int) (string, bool) {
	if at >= 0 && at < len(nativeFun.requiredArgNames) {
		return nativeFun.requiredArgNames[at], true
	} else if at >= len(nativeFun.requiredArgNames) && at < len(nativeFun.requiredArgNames)+len(nativeFun.optionalArgNames) {
		return nativeFun.optionalArgNames[at-len(nativeFun.requiredArgNames)], true
	}
	return "", false
}

func (nativeFun *NativeFun) Call(intp *Interpreter, args map[string]interface{}) (interface{}, error) {
	v, err := nativeFun.fn(args)
	if err != nil {
		return nil, err
	}
	return normalizeValue(v), nil
}

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

func (self *Macro) Required(argNames ...string) *Macro {
	self.requiredArgNames = append(self.requiredArgNames, argNames...)
	return self
}

func (self *Macro) Optional(argNames ...string) *Macro {
	self.optionalArgNames = append(self.optionalArgNames, argNames...)
	return self
}
func (self *Macro) Vararg(argName string) *Macro {
	self.varArgName = argName
	return self
}
func (self *Macro) Help(help string) *Macro {
	self.help = help
	return self
}

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

func (self *Prelude) Load() {
	self.Bind("bind", NewMacro(func(intp *Interpreter, args map[string]Node, varArgs []Node) (any, error) {
		name, err := args["name"].Eval(intp)
		strName, ok := name.(string)
		if !ok {
			return nil, NewErrTypeMismatch("string")
		}
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		intp.Bind(strName, v)
		return v, nil
	}).Required("name", "value").Help("bind value to name in current top scope"))

	self.Bind("set", NewMacro(func(intp *Interpreter, args map[string]Node, varArgs []Node) (any, error) {
		name, err := args["name"].Eval(intp)
		strName, ok := name.(string)
		if !ok {
			return nil, NewErrTypeMismatch("string")
		}
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		if intp.Set(strName, v) {
			return v, nil
		} else {
			intp.Bind(strName, v)
			return v, nil
		}
	}).Required("name", "value").Help("bind value to name in resolved scope, if not found, it's bind to current top scope(the same as 'bind')"))

	self.Bind("block", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
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

	self.Bind("help", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
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

	self.Bind("typeof", NewMacro(func(intp *Interpreter, args map[string]Node, exprlist []Node) (interface{}, error) {
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		return typeName(v), nil
	}).Required("value").Help("the type of a value"))

	installDatetimeFunctions(self)
	installBuiltinFunctions(self)
	installContextFunctions(self)
	installRangeFunctions(self)
}

func (self *Prelude) Bind(name string, value interface{}) *Prelude {
	if _, ok := self.vars[name]; ok {
		panic(fmt.Sprintf("bind(), name '%s' already bound", name))
	}
	self.vars[name] = normalizeValue(value)
	return self
}

func (self *Prelude) Resolve(name string) (interface{}, bool) {
	v, ok := self.vars[name]
	return v, ok
}
