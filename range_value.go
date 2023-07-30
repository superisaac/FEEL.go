package feel

type RangeValue struct {
	StartOpen bool
	Start     any

	EndOpen bool
	End     any
}

func (self RangeValue) BeforePoint(v any) (bool, error) {
	r, err := compareInterfaces(self.End, v)
	if err != nil {
		return false, err
	}
	if self.EndOpen {
		return r >= 0, nil
	} else {
		return r > 0, nil
	}
}

func (self RangeValue) AfterPoint(v any) (bool, error) {
	r, err := compareInterfaces(v, self.Start)
	if err != nil {
		return false, err
	}
	if self.StartOpen {
		return r <= 0, nil
	} else {
		return r < 0, nil
	}
}

func (self RangeValue) BeforeRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.End, other.Start)
	if err != nil {
		return false, err
	}
	if other.StartOpen {
		return r <= 0, nil
	} else {
		return r < 0, nil
	}
}

func (self RangeValue) AfterRange(other RangeValue) (bool, error) {
	r, err := compareInterfaces(self.Start, other.End)
	if err != nil {
		return false, err
	}
	if other.EndOpen {
		return r <= 0, nil
	} else {
		return r < 0, nil
	}
}

func (self RangeValue) Contains(v any) bool {
	switch vv := v.(type) {
	case string:
		strStart, ok := self.Start.(string)
		if !ok {
			return false
		}
		strEnd, ok := self.End.(string)
		if !ok {
			return false
		}
		var startContains = strStart <= vv
		if self.StartOpen {
			startContains = strStart < vv
		}

		var endContains = vv <= strEnd
		if self.EndOpen {
			startContains = vv < strEnd
		}

		return startContains && endContains
	case *Number:
		nStart, ok := self.Start.(*Number)
		if !ok {
			return false
		}
		nEnd, ok := self.End.(*Number)
		if !ok {
			return false
		}
		cmpStart := nStart.Cmp(vv)
		cmpEnd := vv.Cmp(nEnd)

		var startContains = cmpStart <= 0
		if self.StartOpen {
			startContains = cmpStart < 0
		}

		var endContains = cmpEnd <= 0
		if self.EndOpen {
			startContains = cmpEnd < 0
		}
		return startContains && endContains
	}
	return false
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
}
