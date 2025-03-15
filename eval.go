package feel

import (
	"encoding/json"
	"errors"
)

// values

type NullValue struct {
}

func (v NullValue) Equal(other NullValue) bool {
	return true
}

func (v NullValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(nil)
}

var Null = &NullValue{}

func boolValue(condVal any) bool {
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
	case []any:
		return len(v) > 0
	case map[string]any:
		return len(v) > 0
	default:
		return v != nil
	}
}

func typeName(a any) string {
	switch a.(type) {
	case int64:
		return "number"
	case float64:
		return "number"
	case *Number:
		return "number"
	case bool:
		return "bool"
	case string:
		return "string"
	case []any:
		return "list"
	case map[string]any:
		return "context"
	case *NullValue:
		return "null"
	case *FEELDate:
		return "date"
	case *FEELTime:
		return "time"
	case *FEELDatetime:
		return "datetime"
	case *FEELDuration:
		return "duration"
	case *RangeValue:
		return "range"
	case *NativeFun:
		return "function"
	case *FunDef:
		return "function"
	case *Macro:
		return "function"
	default:
		return "unknown"
	}
}

func normalizeValue(v any) any {
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

func (scope Scope) normalizeScope() Scope {
	newScp := make(Scope)
	for key, value := range scope {
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

func (intp Interpreter) String() string {
	return "interpreter"
}

func (intp Interpreter) Len() int {
	return len(intp.ScopeStack)
}

func (intp *Interpreter) Push(scp Scope) {
	intp.ScopeStack = append(intp.ScopeStack, scp.normalizeScope())
}

func (intp *Interpreter) PushEmpty() {
	vars := make(Scope)
	intp.Push(vars)
}

func (intp *Interpreter) Pop() Scope {
	if intp.Len() > 0 {
		top := intp.ScopeStack[len(intp.ScopeStack)-1]
		intp.ScopeStack = intp.ScopeStack[:len(intp.ScopeStack)-1]
		return top
	}
	return nil
}

// resolve a name from the top of scopestack to bottom
func (intp Interpreter) Resolve(name string) (any, bool) {
	for at := len(intp.ScopeStack) - 1; at >= 0; at-- {
		if v, ok := intp.ScopeStack[at][name]; ok {
			return v, true
		}
	}
	if prelude, ok := GetPrelude().Resolve(name); ok {
		return prelude, ok
	}
	return nil, false
}

// resolve the name and set to new value
func (intp Interpreter) Set(name string, value any) bool {
	for at := len(intp.ScopeStack) - 1; at >= 0; at-- {
		if _, ok := intp.ScopeStack[at][name]; ok {
			intp.ScopeStack[at][name] = value
			return true
		}
	}
	return false
}

// bind the value to the name of current scope
func (intp *Interpreter) Bind(name string, value any) {
	if intp.Len() > 0 {
		intp.ScopeStack[intp.Len()-1][name] = normalizeValue(value)
	} else {
		panic("empty stack")
	}
}

// Node's eval functions

// Evaluate Number node
func (n NumberNode) Eval(intp *Interpreter) (any, error) {
	return NewNumber(n.Value), nil
}

// Evaluate bool node
func (node BoolNode) Eval(intp *Interpreter) (any, error) {
	return node.Value, nil
}

func (node NullNode) Eval(intp *Interpreter) (any, error) {
	return Null, nil
}

func (node StringNode) Eval(intp *Interpreter) (any, error) {
	return node.Content(), nil
}

func (node TemporalNode) Eval(intp *Interpreter) (any, error) {
	return ParseTemporalValue(node.Content())
}

func (node Var) Eval(intp *Interpreter) (any, error) {
	if v, ok := intp.Resolve(node.Name); ok {
		return v, nil
	} else {
		//return nil, NewErrKeyNotFound(node.Name)
		return Null, nil
	}
}

func (node RangeNode) Eval(intp *Interpreter) (any, error) {
	startVal, err := node.Start.Eval(intp)
	if err != nil {
		return nil, err
	}
	endVal, err := node.End.Eval(intp)
	if err != nil {
		return nil, err
	}
	return &RangeValue{
		Start:     startVal,
		StartOpen: node.StartOpen,
		End:       endVal,
		EndOpen:   node.EndOpen,
	}, nil
}

func (node ArrayNode) Eval(intp *Interpreter) (any, error) {
	var arr []any
	for _, elem := range node.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func (node ExprList) Eval(intp *Interpreter) (any, error) {
	var finalRet any = nil
	for _, elem := range node.Elements {
		v, err := elem.Eval(intp)
		if err != nil {
			return nil, err
		}
		finalRet = v
	}
	return finalRet, nil
}

func (node MultiTests) Eval(intp *Interpreter) (any, error) {
	for _, elem := range node.Elements {
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

func (node MapNode) Eval(intp *Interpreter) (any, error) {
	mapVal := make(map[string]any)
	for _, item := range node.Values {

		v, err := item.Value.Eval(intp)
		if err != nil {
			return nil, err
		}
		mapVal[item.Name] = v
	}
	return mapVal, nil
}

func (node DotOp) Eval(intp *Interpreter) (any, error) {
	leftVal, err := node.Left.Eval(intp)
	if err != nil {
		return nil, err
	}
	if mapVal, ok := leftVal.(map[string]any); ok {
		if val, found := mapVal[node.Attr]; found {
			return val, nil
		} else {
			return nil, NewErrKeyNotFound(node.Attr)
		}
	} else if obj, ok := leftVal.(HasAttrs); ok {
		if v, found := obj.GetAttr(node.Attr); found {
			return normalizeValue(v), nil
		} else {
			//return nil, NewEvalError(-4001, "attr error", fmt.Sprintf("cannot get attr %s", node.Attr))
			return nil, NewErrKeyNotFound(node.Attr)

		}
	} else {
		return nil, NewErrTypeMismatch("map")
		//return Null, nil
	}
}

func (node IfExpr) Eval(intp *Interpreter) (any, error) {
	condVal, err := node.Cond.Eval(intp)
	if err != nil {
		return nil, err
	}

	if boolValue(condVal) {
		brVal, err := node.ThenBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	} else {
		brVal, err := node.ElseBranch.Eval(intp)
		if err != nil {
			return nil, err
		}
		return brVal, nil
	}
}

func (node ForExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := node.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		results := make([]any, 0)
		for _, val := range aList {
			intp.Bind(node.Varname, val)

			res, err := node.ReturnExpr.Eval(intp)
			if err != nil {
				intp.Pop()
				return nil, err
			}
			results = append(results, res)
		}
		intp.Pop()
		return results, nil
	} else {
		return nil, NewErrTypeMismatch("list")
	}
}

func (node SomeExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := node.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		for _, val := range aList {
			intp.Bind(node.Varname, val)

			res, err := node.FilterExpr.Eval(intp)
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
		return nil, NewErrTypeMismatch("list")
	}
}

func (node EveryExpr) Eval(intp *Interpreter) (any, error) {
	listLike, err := node.ListExpr.Eval(intp)
	if err != nil {
		return nil, err
	}

	if aList, ok := listLike.([]any); ok {
		intp.PushEmpty()
		chooses := make([]any, 0)
		for _, val := range aList {
			intp.Bind(node.Varname, val)

			res, err := node.FilterExpr.Eval(intp)
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
		return nil, NewErrTypeMismatch("list")
	}
}

func (node FunDef) Eval(intp *Interpreter) (any, error) {
	return &FunDef{
		Args: node.Args,
		Body: node.Body,
	}, nil
}

func (node FunDef) EvalCall(intp *Interpreter, args []any) (any, error) {
	if len(args) != len(node.Args) {
		return nil, errors.New("eval call argument size mismatch")
	}
	intp.PushEmpty()
	defer intp.Pop()
	for i, argName := range node.Args {
		intp.Bind(argName, args[i])
	}
	return node.Body.Eval(intp)
}

func (node FunCall) Eval(intp *Interpreter) (any, error) {
	v, err := node.FunRef.Eval(intp)
	if err != nil {
		return nil, err
	}
	switch r := v.(type) {
	case *FunDef:
		return node.EvalFunDef(intp, r)
	case *NativeFun:
		return node.EvalNativeFun(intp, r)
	case *Macro:
		return node.EvalMacro(intp, r)
	default:
		return nil, NewErrTypeMismatch("function")
	}
}

func (node FunCall) EvalNativeFun(intp *Interpreter, funDef *NativeFun) (any, error) {
	argVals := make(map[string]any)
	if node.keywordArgs {
		kwArgMap, err := node.evalArgsToMap(intp)
		if err != nil {
			return nil, err
		}

		for _, argName := range funDef.requiredArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				return nil, NewErrKeywordArgument(argName)
			}
		}

		for _, argName := range funDef.optionalArgNames {
			if v, ok := kwArgMap[argName]; ok {
				argVals[argName] = v
			}
		}
	} else {
		if len(node.Args) < len(funDef.requiredArgNames) {
			required := funDef.requiredArgNames[len(node.Args):len(funDef.requiredArgNames)]
			return nil, NewErrTooFewArguments(required)
		}
		for i, argNode := range node.Args {
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
					varargs := vars.([]any)
					varargs = append(varargs, a)
					argVals[funDef.varArgName] = varargs
				} else {
					argVals[funDef.varArgName] = []any{a}
				}
			} else {
				//return nil, NewEvalError(-5002, "too many arguments")
				return nil, NewErrTooManyArguments()
			}
		}
	}
	return funDef.Call(intp, argVals)
}

func (node FunCall) evalArgsToMap(intp *Interpreter) (map[string]any, error) {
	if !node.keywordArgs {
		return nil, errors.New("funcall has no keyword args")
	}
	kwArgMap := make(map[string]any)
	for _, argNode := range node.Args {
		a, err := argNode.arg.Eval(intp)
		if err != nil {
			return nil, err
		}
		kwArgMap[argNode.argName] = a
	}
	return kwArgMap, nil
}

func (node FunCall) EvalMacro(intp *Interpreter, macro *Macro) (any, error) {
	if len(macro.requiredArgNames) > len(node.Args) {
		return nil, NewErrTooFewArguments(macro.requiredArgNames[len(node.Args):])
		//return nil, NewEvalError(-1005, "number of args of macro mismatch")
	}

	argNodes := make(map[string]Node)
	var varArgs []Node
	if node.keywordArgs {
		kwArgMap := make(map[string]Node)
		for _, argNode := range node.Args {
			kwArgMap[argNode.argName] = argNode.arg
		}

		for _, argName := range macro.requiredArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argNodes[argName] = ast
			} else {
				//return nil, NewEvalError(-5001, "no keyword argument", fmt.Sprintf("no keyword argument %s", argName))
				return nil, NewErrKeywordArgument(argName)
			}
		}

		for _, argName := range macro.optionalArgNames {
			if ast, ok := kwArgMap[argName]; ok {
				argNodes[argName] = ast
			}
		}
	} else {
		if len(node.Args) < len(macro.requiredArgNames) {
			//reqArgs := strings.Join(macro.requiredArgNames[len(node.Args):len(macro.requiredArgNames)], ", ")
			//return nil, NewEvalError(-5003, "too few arguments", fmt.Sprintf("more arguments required: %s", reqArgs))
			return nil, NewErrTooFewArguments(macro.requiredArgNames[len(node.Args):])
		}
		for i, argNode := range node.Args {
			if i < len(macro.requiredArgNames) {
				argNodes[macro.requiredArgNames[i]] = argNode.arg
			} else if i < len(macro.requiredArgNames)+len(macro.optionalArgNames) {
				argNodes[macro.optionalArgNames[i-len(macro.requiredArgNames)]] = argNode.arg
			} else if macro.varArgName != "" {
				varArgs = append(varArgs, argNode.arg)
			} else {
				//return nil, NewEvalError(-5002, "too many arguments")
				return nil, NewErrTooManyArguments()
			}
		}
	}
	return macro.fn(intp, argNodes, varArgs)
}

func (node FunCall) EvalFunDef(intp *Interpreter, funDef *FunDef) (any, error) {
	if len(funDef.Args) > len(node.Args) {
		//return nil, NewEvalError(-1004, "number of args mismatch")
		return nil, NewErrTooFewArguments(funDef.Args[len(node.Args):])
	} else if len(funDef.Args) < len(node.Args) {
		return nil, NewErrTooManyArguments()
	}
	//var args []any
	intp.PushEmpty()
	defer intp.Pop()

	if node.keywordArgs {
		kwArgMap, err := node.evalArgsToMap(intp)
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
		for i, argNode := range node.Args {
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

func EvalString(input string) (any, error) {
	return EvalStringWithScope(input, nil)
}

func EvalStringWithScope(input string, scope Scope) (any, error) {
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
