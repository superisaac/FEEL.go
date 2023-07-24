package feel

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
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

	// string functions
	prelude.BindNativeFunc("string length", func(s string) (int, error) {
		return len(s), nil
	}, []string{"string"})

	type SubstringArgs struct {
		Str      string  `json:"string"`
		StartPos *Number `json:"start position"`
		Length   *Number `json:"length,omitempty"`
	}

	prelude.BindRawNativeFunc("substring", func(kwargs map[string]interface{}) (interface{}, error) {
		args := SubstringArgs{}
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
	}, []string{"string", "start position", "length"})

	prelude.BindNativeFunc("upper case", func(s string) (string, error) {
		return strings.ToUpper(s), nil
	}, []string{"string"})

	prelude.BindNativeFunc("lower case", func(s string) (string, error) {
		return strings.ToLower(s), nil
	}, []string{"string"})

	prelude.BindNativeFunc("contains", func(s string, match string) (bool, error) {
		return strings.Contains(s, match), nil
	}, []string{"string", "match"})

	prelude.BindNativeFunc("starts with", func(s string, match string) (bool, error) {
		return strings.HasPrefix(s, match), nil
	}, []string{"string", "match"})

	prelude.BindNativeFunc("ends with", func(s string, match string) (bool, error) {
		return strings.HasSuffix(s, match), nil
	}, []string{"string", "match"})

	// list functions
	prelude.BindNativeFunc("list contains", func(list []interface{}, elem interface{}) (bool, error) {
		for _, entry := range list {
			if cmp, err := compareInterfaces(entry, elem); err == nil && cmp == 0 {
				return true, err
			}
		}
		return false, nil
	}, []string{"list", "element"})

	prelude.BindNativeFunc("count", func(list []interface{}) (int, error) {
		return len(list), nil
	}, []string{"list"})

	prelude.BindNativeFunc("min", func(list []interface{}) (interface{}, error) {
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
	}, []string{"list"})

	prelude.BindNativeFunc("max", func(list []interface{}) (interface{}, error) {
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
	}, []string{"list"})

	prelude.BindNativeFunc("sum", func(list []interface{}) (*Number, error) {
		sum := ParseNumber(0)
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Add(numEntry)
			}
		}
		return sum, nil
	}, []string{"list"})

	prelude.BindNativeFunc("product", func(list []interface{}) (*Number, error) {
		sum := ParseNumber(1)
		for _, entry := range list {
			if numEntry, ok := entry.(*Number); ok {
				sum = sum.Mul(numEntry)
			}
		}
		return sum, nil
	}, []string{"list"})

	prelude.BindNativeFunc("mean", func(list []interface{}) (*Number, error) {
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
	}, []string{"list"})

	prelude.BindNativeFunc("median", func(list []interface{}) (*Number, error) {
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
			medPos := (len(numberList) / 2) - 1
			return numberList[medPos].Add(numberList[medPos+1]).Mul(ParseNumber("0.5")), nil
		}
	}, []string{"list"})

}
