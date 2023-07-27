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

func (self ArgTypeError) Error() string {
	return fmt.Sprintf("type error at %d, expect %s", self.index, self.expect)
}

// ArgSizeError
type ArgSizeError struct {
	has    int
	expect int
}

func (self ArgSizeError) Error() string {
	return fmt.Sprintf("argument size error, has %d, expect %d", self.has, self.expect)
}

// native function
type NativeFunDef func(args map[string]interface{}) (interface{}, error)

type NativeFun struct {
	fn               NativeFunDef
	requiredArgNames []string
	optionalArgNames []string
	varArgName       string
}

func NewNativeFunc(fn NativeFunDef) *NativeFun {
	return &NativeFun{fn: fn}
}

func (self *NativeFun) Required(argNames ...string) *NativeFun {
	self.requiredArgNames = append(self.requiredArgNames, argNames...)
	return self
}

func (self *NativeFun) Optional(argNames ...string) *NativeFun {
	self.optionalArgNames = append(self.optionalArgNames, argNames...)
	return self
}

func (self *NativeFun) Vararg(argName string) *NativeFun {
	self.varArgName = argName
	return self
}

func (self NativeFun) ArgNameAt(at int) (string, bool) {
	if at >= 0 && at < len(self.requiredArgNames) {
		return self.requiredArgNames[at], true
	} else if at >= len(self.requiredArgNames) && at < len(self.requiredArgNames)+len(self.optionalArgNames) {
		return self.optionalArgNames[at-len(self.requiredArgNames)], true
	}
	return "", false
}

func (self *NativeFun) Call(intp *Interpreter, args map[string]interface{}) (interface{}, error) {
	v, err := self.fn(args)
	if err != nil {
		return nil, err
	}
	return normalizeValue(v), nil
}

// macro
type MacroDef func(intp *Interpreter, args map[string]AST, varArgs []AST) (interface{}, error)
type Macro struct {
	fn               MacroDef
	requiredArgNames []string
	optionalArgNames []string
	varArgName       string
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

func (self *Prelude) Load() {
	self.Bind("bind", NewMacro(func(intp *Interpreter, args map[string]AST, varArgs []AST) (interface{}, error) {
		name, err := args["name"].Eval(intp)
		strName, ok := name.(string)
		if !ok {
			return nil, NewEvalError(-9001, "arg[name].type is not string")
		}
		v, err := args["value"].Eval(intp)
		if err != nil {
			return nil, err
		}
		intp.Bind(strName, v)
		return nil, nil
	}).Required("name", "value"))

	installDatetimeFunctions(self)
	installBuiltinFunctions(self)
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

// buildin native funcs
func nativeBind(intp *Interpreter, varname string, value interface{}) (interface{}, error) {
	intp.Bind(varname, value)
	return nil, nil
}
