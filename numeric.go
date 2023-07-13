package feel

// spec on FEEL's number is https://kiegroup.github.io/dmn-feel-handbook/#number
import (
	"encoding/json"
	"fmt"
	"math/big"
)

const (
	Prec = 34 * 8
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

func ParseNumber(v interface{}) *Number {
	switch vv := v.(type) {
	case int:
		return NewNumberFromInt64(int64(vv))
	case int64:
		return NewNumberFromInt64(vv)
	case float64:
		return NewNumberFromFloat(vv)
	case string:
		return NewNumber(vv)
	default:
		panic(fmt.Sprintf("bad number %#v", v))
	}
}

func (self Number) Int64() int64 {
	i64v, _ := self.v.Int64()
	return i64v
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
