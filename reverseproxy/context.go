package reverseproxy

import "context"

var VarsCtxKey = "data-key"

func InitRequestContext() context.Context {
	var data = map[string]any{"init": true}
	return context.WithValue(context.Background(), VarsCtxKey, data)
}

// GetVar gets a value out of the context's variable table by key.
// If the key does not exist, the return value will be nil.
func GetVar(ctx context.Context, key string) any {
	varMap, ok := ctx.Value(VarsCtxKey).(map[string]any)
	if !ok {
		return nil
	}
	return varMap[key]
}

// SetVar sets a value in the context's variable table with
// the given key. It overwrites any previous value with the
// same key.
//
// If the value is nil (note: non-nil interface with nil
// underlying value does not count) and the key exists in
// the table, the key+value will be deleted from the table.
func SetVar(ctx context.Context, key string, value any) {
	varMap, ok := ctx.Value(VarsCtxKey).(map[string]any)
	if !ok {
		return
	}
	if value == nil {
		if _, ok := varMap[key]; ok {
			delete(varMap, key)
			return
		}
	}
	varMap[key] = value
}
