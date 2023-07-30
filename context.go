package feel

func contextGetByKeys(ctx map[string]interface{}, keys []string) (interface{}, bool) {
	for i, key := range keys {
		if i == len(keys)-1 {
			v, ok := ctx[key]
			return v, ok
		} else {
			v, ok := ctx[key]
			if !ok {
				return nil, false
			}
			if subctx, ok := v.(map[string]interface{}); ok {
				ctx = subctx
			} else {
				return nil, false
			}
		}
	}
	return nil, false
}

func contextProbePut(ctx map[string]interface{}, keys []string) bool {
	for i, key := range keys {
		if i == len(keys)-1 {
			return true
		} else {
			v, ok := ctx[key]
			if !ok {
				// empty cell can be put
				return true
			}
			if subctx, ok := v.(map[string]interface{}); ok {
				ctx = subctx
			} else {
				// sub ctx is not map
				return false
			}
		}
	}
	return false
}

func contextCopy(ctx map[string]interface{}) map[string]interface{} {
	newCtx := make(map[string]interface{})
	for k, v := range ctx {
		newCtx[k] = v
	}
	return newCtx
}

func contextPutKeys(ctx map[string]interface{}, keys []string, value interface{}) (map[string]interface{}, bool) {
	if !contextProbePut(ctx, keys) {
		// cannot put keys
		return ctx, false
	}

	rootCtx := ctx
	for i, key := range keys {
		if i == len(keys)-1 {
			// the last key
			ctx[key] = value
			return rootCtx, true
		} else {
			if v, ok := ctx[key]; ok {
				if subctx, ok := v.(map[string]interface{}); ok {
					ctx = subctx
				} else {
					return rootCtx, false
				}
			} else {
				subctx := make(map[string]interface{})
				ctx[key] = subctx
				ctx = subctx
			}
		}
	}
	return rootCtx, false
}

func installContextFunctions(prelude *Prelude) {
	// context/map functions
	prelude.Bind("get value", NewNativeFunc(func(kwargs map[string]interface{}) (interface{}, error) {
		type getvalueByKey struct {
			Context map[string]interface{} `json:"context"`
			Key     string                 `json:"key"`
		}

		type getvalueByKeys struct {
			Context map[string]interface{} `json:"context"`
			Keys    []string               `json:"key"`
		}

		argsByKey := getvalueByKey{}

		if err := decodeKWArgs(kwargs, &argsByKey); err != nil {
			argsByKeys := getvalueByKeys{}
			if err := decodeKWArgs(kwargs, &argsByKeys); err != nil {
				return nil, err
			}

			if v, ok := contextGetByKeys(argsByKeys.Context, argsByKeys.Keys); ok {
				return v, nil
			} else {
				return Null, nil
			}
		} else {
			if v, ok := argsByKey.Context[argsByKey.Key]; ok {
				return v, nil
			} else {
				return Null, nil
			}
		}
	}).Required("context", "key"))

	prelude.Bind("get entries", wrapTyped(func(ctx map[string]interface{}) ([](map[string]interface{}), error) {
		entries := make([](map[string]interface{}), 0)
		for k, v := range ctx {
			ent := map[string]interface{}{
				"key":   k,
				"value": v,
			}
			entries = append(entries, ent)
		}
		return entries, nil
	}).Required("context"))

	prelude.Bind("context put", NewNativeFunc(func(kwargs map[string]interface{}) (interface{}, error) {
		type putByKey struct {
			Context map[string]interface{} `json:"context"`
			Key     string                 `json:"key"`
			Value   interface{}            `json:"value"`
		}

		type putByKeys struct {
			Context map[string]interface{} `json:"context"`
			Keys    []string               `json:"key"`
			Value   interface{}            `json:"value"`
		}

		argsByKey := putByKey{}

		if err := decodeKWArgs(kwargs, &argsByKey); err != nil {
			argsByKeys := putByKeys{}
			if err := decodeKWArgs(kwargs, &argsByKeys); err != nil {
				return nil, err
			}
			ctx, _ := contextPutKeys(argsByKeys.Context, argsByKeys.Keys, argsByKeys.Value)
			return ctx, nil
		} else {
			ctx, _ := contextPutKeys(argsByKey.Context, []string{argsByKey.Key}, argsByKey.Value)
			return ctx, nil
		}
	}).Required("context", "key", "value"))

}
