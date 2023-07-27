package feel

import (
	"errors"
	"fmt"
	"strings"
)

type EvalError struct {
	Code    int
	Short   string
	Message string
}

func (self EvalError) Error() string {
	return fmt.Sprintf("%d %s %s", self.Code, self.Short, self.Message)
}

// values

type NullValue struct {
}

func (self NullValue) Equal(other NullValue) bool {
	return true
}

var Null = &NullValue{}

func boolValue(condVal interface{}) bool {
	switch v := condVal.(type) {
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case *Number:
		return !v.Equal(*Zero)
	case bool:
		return v
	case string:
		return v != ""
	case []interface{}:
		return v != nil && len(v) > 0
	default:
		return v != nil
	}
}

func normalizeValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case int:
		return NewNumberFromInt64(int64(vv))
	case int64:
		return NewNumberFromInt64(vv)
	case float64:
		return NewNumberFromFloat(vv)
	default:
		return vv
	}
}

func (self Scope) normalizeScope() Scope {
	newScp := make(Scope)
	for key, value := range self {
		newScp[key] = normalizeValue(value)
	}
	return newScp
}

// intepreter
func NewIntepreter() *Interpreter {
	intp := &Interpreter{}
	intp.PushEmpty()
	return intp
}

func (self Interpreter) String() string {
	return "interpreter"
}

func (self Interpreter) Len() int {
	return len(self.ScopeStack)
}

func (self *Interpreter) Push(scp Scope) {
	self.ScopeStack = append(self.ScopeStack, scp.normalizeScope())
}

func (self *Interpreter) PushEmpty() {
	vars := make(Scope)
	self.Push(vars)
}

func (self *Interpreter) Pop() Scope {
	if self.Len() > 0 {
		top := self.ScopeStack[len(self.ScopeStack)-1]
		self.ScopeStack = self.ScopeStack[:len(self.ScopeStack)-1]
		return top
	}
	return nil
}

func (self Interpreter) Resolve(name string) (interface{}, bool) {
	for at := len(self.ScopeStack) - 1; at >= 0; at-- {
		if v, ok := self.ScopeStack[at][name]; ok {
			return v, true
		}
	}
	if prelude, ok := GetPrelude().Resolve(name); ok {
		return prelude, ok
	}

	return nil, false
}

func (self *Interpreter) Bind(name string, value interface{}) {
	if self.Len() > 0 {
		self.ScopeStack[self.Len()-1][name] = normalizeValue(value)
	} else {
		panic("empty stack")
	}
}

func NewEvalError(code int, short string, msgs ...string) *EvalError {
	message := strings.Join(msgs, " ")
	return &EvalError{
		Code:    code,
		Short:   short,
		Message: message,
	}
}

// AST's eval functions
func (self NumberNode) Eval(intp *Interpreter) (interface{}, error) {
	return NewNumber(self.Value), nil
}

func (self BoolNode) Eval(intp *Interpreter) (interface{}, error) {
	return self.Value, nil
}

func (self NullNode) Eval(intp *Interpreter) (interface{}, error) {
	return Null, nil
}

func (self StringNode) Eval(intp *Interpreter) (interface{}, error) {
	return self.Content(), nil
}

func (self TemporalNode) Eval(intp *Interpreter) (interface{}, error) {
	return ParseTemporalValue(self.Content())
}

func (self Var) Eval(intp *Interpreter) (interface{}, error) {
	if v, ok := intp.Resolve(self.Name); ok {
		return v, nil
	} else {
		//return nil, NewEvalError(-1000, "fail to resolve name", fmt.Sprintf("fail to resolve name %s", self.Name))
		return Null, nil
	}
}

func (self RangeNode) Eval(intp *Interpreter) (interface{}, error) {
	startVal, err := self.Start.Eval(intp)
	if err != nil {
		return nil, err
	}
	endVal, err := self.End.Eval(intp)
	if err != nil {
		return nil, err
	}
	return &RangeValue{
		Start:     startVal,
		StartOpen: self.StartOpen,
		End:       endVal,
		EndOpen:   self.EndOpen,
	}, nil
}

func (self ArrayNode) Eval(intp *Interpreter) (interface{}, error) {
	var arr []interface{}
	for _, elem := range self.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func (self ExprList) Eval(intp *Interpreter) (interface{}, error) {
	var finalRet interface{} = nil
	for _, elem := range self.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		finalRet = v
	}
	return finalRet, nil
}

func (self MultiTests) Eval(intp *Interpreter) (interface{}, error) {
	for _, elem := range self.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		if boolValue(v) {
			return true, nil
		}
	}
	return false, nil
}

func (self MapNode) Eval(intp *Interpreter) (interface{}, error) {
	mapVal := make(map[string]interface{})
	for _, item := range self.Values {

		v, err := item.Value.Eval(intp)
		if err != nil {
			return nil, err
		}
		mapVal[item.Name] = v
	}
	return mapVal, nil
}

func (self DotOp) Eval(intp *Interpreter) (interface{}, error) {
	leftVal, err := self.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	if mapVal, ok := leftVal.(map[string]interface{}); ok {
		if val, found := mapVal[self.Attr]; found {
			return val, nil
		} else {
			return nil, NewEvalError(-4000, "map key error", fmt.Sprintf("cannot find map attribute %s", self.Attr))
		}
	} else if obj, ok := leftVal.(HasAttrs); ok {
		if v, ok := obj.GetAttr(self.Attr); ok {
			return normalizeValue(v), nil
		} else {
			return nil, NewEvalError(-4001, "attr error", fmt.Sprintf("cannot get attr %s", self.Attr))
		}
	} else {
		//return nil, NewEvalError(-4001, "type mismatch", "is not map")
		return Null, nil
	}
}

func (self IfExpr) Eval(intp *Interpreter) (interface{}, error) {
	condVal, err := self.Cond.Eval(intp)
	if err != nil {
		return nil, err
	}

	if boolValue(condVal) {
		brVal, err := self.ThenBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	} else {
		brVal, err := self.ElseBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	}
}

func (self ForExpr) Eval(intp *Interpreter) (interface{}, error) {
	listLike, err := self.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]interface{}); ok {
		intp.PushEmpty()
		results := make([]interface{}, 0)
		for _, val := range aList {
			intp.Bind(self.Varname, val)

			res, err := self.ReturnExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}
			results = append(results, res)
		}
		intp.Pop()
		return results, nil
	} else {
		return nil, NewEvalError(-6000, "not a list")
	}
}

func (self SomeExpr) Eval(intp *Interpreter) (interface{}, error) {
	listLike, err := self.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]interface{}); ok {
		intp.PushEmpty()
		for _, val := range aList {
			intp.Bind(self.Varname, val)

			res, err := self.FilterExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}
			if boolValue(res) {
				return val, nil
			}
		}
		intp.Pop()
		return nil, nil
	} else {
		return nil, NewEvalError(-6000, "not a list")
	}
}

func (self EveryExpr) Eval(intp *Interpreter) (interface{}, error) {
	listLike, err := self.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]interface{}); ok {
		intp.PushEmpty()
		chooses := make([]interface{}, 0)
		for _, val := range aList {
			intp.Bind(self.Varname, val)

			res, err := self.FilterExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}

			if boolValue(res) {
				chooses = append(chooses, val)
			}
		}
		intp.Pop()
		return chooses, nil
	} else {
		return nil, NewEvalError(-6000, "not a list")
	}
}

func (self FunDef) Eval(intp *Interpreter) (interface{}, error) {
	return &FunDef{
		Args: self.Args,
		Body: self.Body,
	}, nil
}

func (self FunCall) Eval(intp *Interpreter) (interface{}, error) {
	v, err := self.FunRef.Eval(intp)
	if err != nil {
		return nil, err
	}
	switch r := v.(type) {
	case *FunDef:
		return self.EvalFunDef(intp, r)
	case *NativeFun:
		return self.EvalNativeFun(intp, r)
	case *Macro:
		return self.EvalMacro(intp, r)
	default:
		return nil, NewEvalError(-1003, "call on a non function")
	}
}

func (self FunCall) EvalNativeFun(intp *Interpreter, funDef *NativeFun) (interface{}, error) {
	argVals := make(map[string]interface{})
	if self.keywordArgs {
		kwArgMap, err := self.evalArgsToMap(intp)
		if err != nil {
			return nil, err
		}

		for _, argName := range funDef.requiredArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			} else {
				return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
			}
		}

		for _, argName := range funDef.optionalArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			}
		}
	} else {
		if len(self.Args) < len(funDef.requiredArgNames) {
			reqArgs := strings.Join(funDef.requiredArgNames[len(self.Args):len(funDef.requiredArgNames)], ", ")
			return nil, NewEvalError(-5003, "too few arguments", fmt.Sprintf("more arguments required: %s", reqArgs))
		}
		for i, argNode := range self.Args {
			a, err := argNode.arg.Eval(intp)
			if err != nil {
				return nil, err
			}
			if i < len(funDef.requiredArgNames) {
				argVals[funDef.requiredArgNames[i]] = a
			} else if i < len(funDef.requiredArgNames)+len(funDef.optionalArgNames) {
				argVals[funDef.optionalArgNames[i-len(funDef.requiredArgNames)]] = a
			} else if funDef.varArgName != "" {
				if vars, ok := argVals[funDef.varArgName]; ok {
					varargs := vars.([]interface{})
					varargs = append(varargs, a)
					argVals[funDef.varArgName] = varargs
				} else {
					argVals[funDef.varArgName] = []interface{}{a}
				}
			} else {
				return nil, NewEvalError(-5002, "too many arguments")
			}
		}
	}
	return funDef.Call(intp, argVals)
}

func (self FunCall) evalArgsToMap(intp *Interpreter) (map[string]interface{}, error) {
	if !self.keywordArgs {
		return nil, errors.New("funcall has no keyword args")
	}
	kwArgMap := make(map[string]interface{})
	for _, argNode := range self.Args {
		a, err := argNode.arg.Eval(intp)
		if err != nil {
			return nil, err
		}
		kwArgMap[argNode.argName] = a
	}
	return kwArgMap, nil
}

func (self FunCall) EvalMacro(intp *Interpreter, macro *Macro) (interface{}, error) {
	if len(macro.requiredArgNames) > len(self.Args) {
		return nil, NewEvalError(-1005, "number of args of macro mismatch")
	}

	argASTs := make(map[string]AST)
	var varArgs []AST
	if self.keywordArgs {
		kwArgMap := make(map[string]AST)
		for _, argNode := range self.Args {
			kwArgMap[argNode.argName] = argNode.arg
		}

		for _, argName := range macro.requiredArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argASTs[argName] = ast
			} else {
				return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
			}
		}

		for _, argName := range macro.optionalArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argASTs[argName] = ast
			}
		}
	} else {
		if len(self.Args) < len(macro.requiredArgNames) {
			reqArgs := strings.Join(macro.requiredArgNames[len(self.Args):len(macro.requiredArgNames)], ", ")
			return nil, NewEvalError(-5003, "too few arguments", fmt.Sprintf("more arguments required: %s", reqArgs))
		}
		for i, argNode := range self.Args {
			if i < len(macro.requiredArgNames) {
				argASTs[macro.requiredArgNames[i]] = argNode.arg
			} else if i < len(macro.requiredArgNames)+len(macro.optionalArgNames) {
				argASTs[macro.optionalArgNames[i-len(macro.requiredArgNames)]] = argNode.arg
			} else if macro.varArgName != "" {
				varArgs = append(varArgs, argNode.arg)
			} else {
				return nil, NewEvalError(-5002, "too many arguments")
			}
		}
	}
	return macro.fn(intp, argASTs, varArgs)
}

func (self FunCall) EvalFunDef(intp *Interpreter, funDef *FunDef) (interface{}, error) {
	if len(funDef.Args) != len(self.Args) {
		return nil, NewEvalError(-1004, "number of args mismatch")
	}
	//var args []interface{}
	intp.PushEmpty()
	defer intp.Pop()

	if self.keywordArgs {
		kwArgMap, err := self.evalArgsToMap(intp)
		if err != nil {
			return nil, err
		}

		for _, argName := range funDef.Args {
			if v, ok := kwArgMap[argName]; ok {
				//argVals = append(argVals, v)
				intp.Bind(argName, v)
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				intp.Bind(argName, Null)
			}
		}
	} else {
		for i, argNode := range self.Args {
			a, err := argNode.arg.Eval(intp)
			if err != nil {
				return nil, err
			}
			intp.Bind(funDef.Args[i], a)
		}
	}
	ret, err := funDef.Body.Eval(intp)
	return ret, err
}

func EvalString(input string) (interface{}, error) {
	return EvalStringWithScope(input, nil)
}

func EvalStringWithScope(input string, scope Scope) (interface{}, error) {
	ast, err := ParseString(input)
	if err != nil {
		return nil, err
	}
	intp := NewIntepreter()
	if scope != nil {
		intp.Push(scope)
	}
	r, err := ast.Eval(intp)
	return r, err
}
