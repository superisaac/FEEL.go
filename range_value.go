package feel

import (
//"fmt"
)

type RangeValue struct {
	StartOpen bool
	Start     any

	EndOpen bool
	End     any
}

func (self RangeValue) BeforePoint(p any) (bool, error) {
	pos, err := self.Position(p)
	if err != nil {
		return false, err
	}
	return pos > 0, nil

}

func (self RangeValue) AfterPoint(p any) (bool, error) {
	pos, err := self.Position(p)
	if err != nil {
		return false, err
	}
	return pos < 0, nil
}

func (self RangeValue) BeforeRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.End, other.Start)
	if err != nil {
		return false, err
	}

	if !self.EndOpen && !other.StartOpen {
		// two ranges meet
		return r < 0, nil
	} else {
		return r <= 0, nil
	}
}

func (self RangeValue) AfterRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.Start, other.End)
	if err != nil {
		return false, err
	}
	if !self.StartOpen && !other.EndOpen {
		return r > 0, nil
	} else {
		return r >= 0, nil
	}
}

func (self RangeValue) Position(p any) (int, error) {
	cmpStart, err := compareInterfaces(p, self.Start)
	if err != nil {
		return 0, err
	}
	if self.StartOpen {
		if cmpStart <= 0 {
			return -1, nil
		}
	} else {
		if cmpStart == 0 {
			return 0, nil
		} else if cmpStart < 0 {
			return -1, nil
		}
	}

	cmpEnd, err := compareInterfaces(p, self.End)
	if err != nil {
		return 0, err
	}
	if self.EndOpen && cmpEnd >= 0 {
		return 1, nil
	} else if !self.EndOpen && cmpEnd > 0 {
		return 1, nil
	}
	return 0, nil

}

func (self RangeValue) Contains(p any) bool {
	r, err := self.Position(p)
	if err != nil {
		panic(err)
	}
	return r == 0
}

func installRangeFunctions(prelude *Prelude) {
	prelude.Bind("before", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		a := kwargs["a"]
		b := kwargs["b"]
		switch va := a.(type) {
		case *RangeValue:
			switch vb := b.(type) {
			case *RangeValue:
				return va.BeforeRange(*vb)
			default:
				return va.BeforePoint(vb)
			}
		default:
			switch vb := b.(type) {
			case *RangeValue:
				return vb.AfterPoint(va)
			default:
				cmp, err := compareInterfaces(va, vb)
				if err != nil {
					return nil, err
				} else {
					return cmp < 0, nil
				}
			}
		}
	}).Required("a", "b"))

	prelude.Bind("after", NewNativeFunc(func(kwargs map[string]any) (any, error) {
		a := kwargs["a"]
		b := kwargs["b"]
		switch va := a.(type) {
		case *RangeValue:
			switch vb := b.(type) {
			case *RangeValue:
				return va.AfterRange(*vb)
			default:
				return va.AfterPoint(vb)
			}
		default:
			switch vb := b.(type) {
			case *RangeValue:
				return vb.BeforePoint(va)
			default:
				cmp, err := compareInterfaces(va, vb)
				if err != nil {
					return nil, err
				} else {
					return cmp > 0, nil
				}
			}
		}
	}).Required("a", "b"))

	prelude.Bind("meets", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		r, err := compareInterfaces(a.End, b.Start)
		if err != nil {
			return false, err
		}
		return !a.EndOpen && !b.StartOpen && r == 0, nil
	}).Required("a", "b"))

	prelude.Bind("met by", wrapTyped(func(a *RangeValue, b *RangeValue) (bool, error) {
		r, err := compareInterfaces(a.Start, b.End)
		if err != nil {
			return false, err
		}
		return !b.EndOpen && !a.StartOpen && r == 0, nil
	}).Required("a", "b"))
}
