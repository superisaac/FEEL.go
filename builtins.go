package feel

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"sort"
	"strings"
)

func decodeKWArgs(input map[string]interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		TagName:  "json",
		Result:   &output,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func extractList(args map[string]interface{}, argName string) ([]interface{}, error) {
	if v, ok := args[argName]; ok {
		if listV, ok := v.([]interface{}); ok {
			if len(listV) == 1 {
				if aList, ok := listV[0].([]interface{}); ok {
					return aList, nil
				}
			}
			return listV, nil
		} else {
			return nil, NewEvalError(-7002, "cannot extract list")
		}
	} else {
		return nil, NewEvalError(-7001, "no argument name")
	}
}

func installBuiltinFunctions(prelude *Prelude) {
	// conversion functions
	prelude.Bind("string", wrapTyped(func(v interface{}) (string, error) {
		return fmt.Sprintf("%s", v), nil
	}).Required("from"))

	prelude.Bind("number", wrapTyped(func(v interface{}) (*Number, error) {
		return ParseNumberWithErr(v)
	}).Required("from"))

	// boolean functions
	prelude.Bind("not", wrapTyped(func(v interface{}) (bool, error) {
		return !boolValue(v), nil
	}).Required("from"))

	prelude.Bind("is defined", NewMacro(func(intp *Interpreter, args []AST) (interface{}, error) {
		if varNode, ok := args[0].(*Var); ok {
			if _, ok := intp.Resolve(varNode.Name); !ok {
				return false, nil
			}
		}
		// TODO: more condition tests
		return true, nil
	}).Required("value"))

	// string functions
	prelude.Bind("string length", wrapTyped(func(s string) (int, error) {
		return len(s), nil
	}).Required("string"))

	prelude.Bind("substring", NewNativeFunc(func(kwargs map[string]interface{}) (interface{}, error) {
		type substringArgs struct {
			Str      string  `json:"string"`
			StartPos *Number `json:"start position"`
			Length   *Number `json:"length,omitempty"`
		}

		args := substringArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		startPos := int(args.StartPos.Int64())
		if startPos >= len(args.Str) {
			return "", nil
		}
		endPos := len(args.Str)
		if args.Length != nil {
			endPos = startPos + int(args.Length.Int64())
			if endPos >= len(args.Str) {
				endPos = len(args.Str)
			}
		}
		subs := args.Str[startPos:endPos]
		return subs, nil
	}).Required("string", "start position").Optional("length"))

	prelude.Bind("upper case", wrapTyped(func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}).Required("string"))

	prelude.Bind("lower case", wrapTyped(func(s string) (string, error) {
		return strings.ToLower(s), nil
	}).Required("string"))

	prelude.Bind("contains", wrapTyped(func(s string, match string) (bool, error) {
		return strings.Contains(s, match), nil
	}).Required("string", "match"))

	prelude.Bind("starts with", wrapTyped(func(s string, match string) (bool, error) {
		return strings.HasPrefix(s, match), nil
	}).Required("string", "match"))

	prelude.Bind("ends with", wrapTyped(func(s string, match string) (bool, error) {
		return strings.HasSuffix(s, match), nil
	}).Required("string", "match"))

	// list functions
	prelude.Bind("list contains", wrapTyped(func(list []interface{}, elem interface{}) (bool, error) {
		for _, entry := range list {
			if cmp, err := compareInterfaces(entry, elem); err == nil && cmp == 0 {
				return true, err
			}
		}
		return false, nil
	}).Required("list", "element"))

	prelude.Bind("count", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		return len(list), nil
	}).Vararg("list"))

	prelude.Bind("min", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var minValue interface{} = nil
		for _, entry := range list {
			if minValue == nil {
				minValue = entry
			} else if cmp, err := compareInterfaces(minValue, entry); err == nil && cmp == 1 {
				// minValue > entry
				minValue = entry
			}
		}
		return minValue, nil
	}).Vararg("list"))

	prelude.Bind("max", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var maxValue interface{} = nil
		for _, entry := range list {
			if maxValue == nil {
				maxValue = entry
			} else if cmp, err := compareInterfaces(maxValue, entry); err == nil && cmp == -1 {
				// maxValue < entry
				maxValue = entry
			}
		}
		return maxValue, nil
	}).Vararg("list"))

	prelude.Bind("sum", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := ParseNumber(0)
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Add(numEntry)
			}
		}
		return sum, nil
	}).Vararg("list"))

	prelude.Bind("product", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := ParseNumber(1)
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Mul(numEntry)
			}
		}
		return sum, nil
	}).Vararg("list"))

	prelude.Bind("mean", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		sum := ParseNumber(0)
		cnt := 0
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Add(numEntry)
				cnt++
			}
		}
		r := sum.FloatDiv(ParseNumber(cnt))
		return r, nil
	}).Vararg("list"))

	prelude.Bind("median", NewNativeFunc(func(args map[string]interface{}) (interface{}, error) {
		list, err := extractList(args, "list")
		if err != nil {
			return nil, err
		}
		var numberList []*Number

		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				numberList = append(numberList, numEntry)
			}
		}
		if len(numberList) == 0 {
			return nil, nil
		} else if len(numberList) == 1 {
			return numberList[0], nil
		}

		sort.Slice(numberList, func(i, j int) bool {
			return numberList[i].Compare(*numberList[j]) == -1
		})

		if len(numberList)%2 == 1 {
			return numberList[len(numberList)/2], nil
		} else {
			medPos := (len(numberList) / 2)
			return numberList[medPos+1].Add(numberList[medPos]).Mul(ParseNumber("0.5")), nil
		}
	}).Vararg("list"))

	prelude.Bind("sublist", NewNativeFunc(func(kwargs map[string]interface{}) (interface{}, error) {
		type sublistArgs struct {
			List     []interface{} `json:"list"`
			StartPos *Number       `json:"start position"`
			Length   *Number       `json:"length,omitempty"`
		}

		args := sublistArgs{}
		if err := decodeKWArgs(kwargs, &args); err != nil {
			return nil, err
		}
		startPos := int(args.StartPos.Int64())
		if startPos >= len(args.List) {
			return "", nil
		}
		endPos := len(args.List)
		if args.Length != nil {
			endPos = startPos + int(args.Length.Int64())
			if endPos >= len(args.List) {
				endPos = len(args.List)
			}
		}
		subs := args.List[startPos:endPos]
		return subs, nil
	}).Required("list", "start position").Optional("length"))
}
