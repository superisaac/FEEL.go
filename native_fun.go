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

type NativeFunDef func(intp *Interpreter, args []interface{}) (interface{}, error)

type NativeFun struct {
	fn       NativeFunDef
	argNames []string
}

func NewNativeFunc(fn NativeFunDef) *NativeFun {
	return &NativeFun{fn: fn}
}

func (self *NativeFun) Call(intp *Interpreter, args []interface{}) (interface{}, error) {
	return self.fn(intp, args)
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
	self.BindNativeFunc("not", nativeNot, "v")
	self.BindNativeFunc("bind", nativeBind, "varname", "value")
	installDateTimeFunctions(self)
}

func (self *Prelude) Bind(name string, value interface{}) {
	self.vars[name] = normalizeValue(value)
}

func (self *Prelude) BindNativeFunc(name string, fn interface{}, argNames ...string) {
	if isdup, argName := hasDupName(argNames); isdup {
		panic(fmt.Sprintf("nativee function %s has duplicate arg name %s", name, argName))
	}
	self.Bind(name, &NativeFun{fn: wrapTyped(fn, argNames), argNames: argNames})
}

func (self *Prelude) Resolve(name string) (interface{}, bool) {
	v, ok := self.vars[name]
	return v, ok
}

// buildin native funcs
func nativeNot(intp *Interpreter, v interface{}) (bool, error) {
	return !boolValue(v), nil
}

func nativeBind(intp *Interpreter, varname string, value interface{}) (interface{}, error) {
	intp.Bind(varname, value)
	return nil, nil
}
