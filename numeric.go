package feel

// spec on FEEL's number is https://kiegroup.github.io/dmn-feel-handbook/#number
import (
	"encoding/json"
	"errors"
	"math/big"
)

const (
	Prec = 34 * 8
)

var (
	ErrParseNumber = errors.New("fail to parse number")
)

type Number struct {
	v *big.Float
}

func NewNumber(strn string) *Number {
	v := new(big.Float)
	v.SetPrec(Prec).SetString(strn)
	return &Number{v: v}
}

func NewNumberFromInt64(input int64) *Number {
	v := new(big.Float)
	v.SetPrec(200).SetInt64(input)
	return &Number{v: v}
}

func NewNumberFromFloat(input float64) *Number {
	v := new(big.Float)
	v.SetPrec(200).SetFloat64(input)
	return &Number{v: v}
}

func ParseNumberWithErr(v interface{}) (*Number, error) {
	switch vv := v.(type) {
	case int:
		return NewNumberFromInt64(int64(vv)), nil
	case int64:
		return NewNumberFromInt64(vv), nil
	case float64:
		return NewNumberFromFloat(vv), nil
	case string:
		return NewNumber(vv), nil
	case *Number:
		return vv, nil
	default:
		return nil, ErrParseNumber
	}
}

func ParseNumber(v interface{}) *Number {
	n, err := ParseNumberWithErr(v)
	if err != nil {
		panic(err)
	}
	return n
}

func (self Number) Int64() int64 {
	i64v, _ := self.v.Int64()
	return i64v
}

func (self Number) Int() int {
	return int(self.Int64())
}

func (self Number) Float64() float64 {
	f64v, _ := self.v.Float64()
	return f64v
}

func (self *Number) Add(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Add(self.v, other.v)
	return &Number{v: newv}
}

func (self *Number) Sub(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Sub(self.v, other.v)
	return &Number{v: newv}
}

func (self *Number) Mul(other *Number) *Number {
	newv := new(big.Float)
	newv.SetPrec(Prec).Mul(self.v, other.v)
	return &Number{v: newv}
}

func (self *Number) Cmp(other *Number) int {
	return self.v.Cmp(other.v)
}

func (self *Number) IntDiv(other *Number) *Number {
	newv := new(big.Int)
	a, _ := self.v.Int(nil)
	b, _ := other.v.Int(nil)
	newv.Div(a, b)
	newf := new(big.Float)
	newf.SetPrec(Prec).SetInt(newv)
	return &Number{v: newf}
}

func (self *Number) FloatDiv(other *Number) *Number {
	a, _ := self.v.Float64()
	b, _ := other.v.Float64()
	newf := new(big.Float)
	newf.SetPrec(Prec).SetFloat64(a / b)
	return &Number{v: newf}
}

func (self *Number) IntMod(other *Number) *Number {
	newv := new(big.Int)
	a, _ := self.v.Int(nil)
	b, _ := other.v.Int(nil)
	newv.Mod(a, b)
	newf := new(big.Float)
	newf.SetPrec(Prec).SetInt(newv)
	return &Number{v: newf}
}

func (self Number) Equal(other Number) bool {
	return self.Compare(other) == 0
}

func (self Number) Compare(other Number) int {
	return self.v.Cmp(other.v)
}

func (self Number) String() string {
	//return self.v.String()
	return self.v.Text('f', 18)
}

func (self Number) MarshalJSON() ([]byte, error) {
	if f32v, acc := self.v.Float32(); acc == big.Exact {
		return json.Marshal(f32v)
	}
	return json.Marshal(self.String())
}

var Zero = ParseNumber(0)
