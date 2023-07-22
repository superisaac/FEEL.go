package feel

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"reflect"
)

func typeIsStruct(tp reflect.Type) bool {
	return (tp.Kind() == reflect.Struct ||
		(tp.Kind() == reflect.Ptr && typeIsStruct(tp.Elem())))
}

func interfaceToValue(a interface{}, outputType reflect.Type) (reflect.Value, error) {
	output := reflect.Zero(outputType).Interface()
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		TagName:  "json",
		Result:   &output,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return reflect.Value{}, err
	}
	err = decoder.Decode(a)
	if err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(output), nil
}

func valueToInterface(tp reflect.Type, val reflect.Value) (interface{}, error) {
	var output interface{}
	//if typeIsStruct(tp) {
	if false {
		output = make(map[string]interface{})
	} else {
		output = reflect.Zero(tp).Interface()
	}
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		TagName:  "json",
		Result:   &output,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(val.Interface())
	if err != nil {
		return nil, err
	}
	return output, nil
}

func wrapTyped(tfunc interface{}, argNames []string) NativeFunDef {

	funcType := reflect.TypeOf(tfunc)
	if funcType.Kind() != reflect.Func {
		panic("tfunc is not func type")
	}

	firstArgNum := 0
	numIn := funcType.NumIn()

	// if false {
	// 	firstArgNum = 1
	// 	// check inputs and 1st argument
	// 	if numIn < firstArgNum {
	// 		panic(errors.New("func must have 1 more arguments"))
	// 	}
	// 	firstArgType := funcType.In(0)
	// 	// TODO: using more methods to check type equal
	// 	if !(firstArgType.Kind() == reflect.Ptr && firstArgType.Elem().Name() == "Interpreter") {
	// 		panic(fmt.Sprintf("the first arg must be %s, it's %s instead", firstArgSpec.String(), firstArgType.Elem().Name()))
	// 	}
	// }

	if numIn != len(argNames)+firstArgNum {
		panic(fmt.Sprintf("arg number msmatch, %d expected, but %d given", numIn-1, len(argNames)))
	}

	// check outputs
	if funcType.NumOut() != 2 {
		panic("func return number must be 2")
	}

	errType := funcType.Out(1)
	errInterface := reflect.TypeOf((*error)(nil)).Elem()

	if !errType.Implements(errInterface) {
		panic("second output does not implement error")
	}

	handler := func(args []interface{}) (interface{}, error) {
		// check inputs
		if numIn > len(args)+firstArgNum {
			return nil, errors.New("no enough params size")
		}

		// params -> []reflect.Value
		fnArgs := []reflect.Value{}
		//fnArgs = append(fnArgs, reflect.ValueOf(intp))
		j := 0
		for i := firstArgNum; i < numIn; i++ {
			argType := funcType.In(i)
			param := args[j]
			j++

			argValue, err := interfaceToValue(param, argType)
			if err != nil {
				return nil, errors.New(
					fmt.Sprintf("arguments %d %s", i+1, err))
			}
			fnArgs = append(fnArgs, argValue)

		}

		// wrap result
		resValues := reflect.ValueOf(tfunc).Call(fnArgs)
		resType := funcType.Out(0)
		errRes := resValues[1].Interface()
		if errRes != nil {
			if err, ok := errRes.(error); ok {
				return nil, err
			} else {
				return nil, errors.New(fmt.Sprintf("error return is not error %+v", errRes))
			}
		}

		res, err := valueToInterface(
			resType, resValues[0])
		return res, err
	}

	return handler
}
