package template

import (
	"strconv"
	"wispy-core/pkg/common"
)

type DataAdapter struct {
	Prefix   string                                                   // Prefix for the data adapter, e.g., "user."
	Writable bool                                                     // Whether the data adapter allows writing
	GetData  func(e *Engine, pathKeys ...string) (value any, ok bool) // Function to get data by path keys
	SetData  func(e *Engine, value any, pathKeys ...string)           // Function to set data by path keys
}

func (e *Engine) ResolveValue(key string, pathKeys ...string) (value any, ok bool) {
	if len(key) == 0 {
		return nil, false
	}

	localCtxValue, localCtxOk := e.GetLocalDataContextValue(key)
	common.Debug("Resolving value for key: %s with path keys: %v", key, pathKeys)

	if localCtxOk {
		if len(pathKeys) > 0 {
			for _, key := range pathKeys {
				newValue, nestedOk := getNestedValue(localCtxValue, key)
				if nestedOk {
					value = newValue // Update to the next level in the context
				} else {
					common.Warning("Local context value not found for key: %s", key)
					return nil, false // If any key is not found, return false
				}
			}
			return value, ok // Return the resolved value from local context
		} else {
			common.Debug("Found value in local context for key: %s", key)
			return localCtxValue, true // Return the value from local context
		}
	}

	adapter, adapterOk := e.GetDataAdapter(pathKeys[0])
	if adapterOk {
		value, ok = adapter.GetData(e, append([]string{key}, pathKeys...)...)
		if !ok {
			common.Warning("Data adapter %s did not return a value for keys: %v", key, pathKeys)
			return nil, false
		} else {
			return value, true
		}
	} else {
		// If the first key matches a data adapter, use it to resolve the value
		common.Debug("No data adapter found for key: %s", key)
		return nil, false
	}
}

func getNestedValue(current any, pathKeys ...string) (value any, ok bool) {
	value = current
	for _, key := range pathKeys {
		switch v := value.(type) {
		case map[string]any:
			value, ok = v[key]
			if !ok {
				return nil, false
			}
		case []any:
			index, err := strconv.Atoi(key)
			if err != nil || index < 0 || index >= len(v) {
				return nil, false
			}
			value = v[index]
		default:
			return nil, false
		}
	}
	return value, true
}
