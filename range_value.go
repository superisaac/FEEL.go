package feel

type RangeValue struct {
	StartOpen bool
	Start     interface{}

	EndOpen bool
	End     interface{}
}

func (self RangeValue) Contains(v interface{}) bool {
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
