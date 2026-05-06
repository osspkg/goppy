package xc

import "context"

func GetValue[T any](ctx context.Context, key any) (value T, ok bool) {
	v := ctx.Value(key)
	if v == nil {
		return
	}
	value, ok = v.(T)
	return
}

func SetValue[T any](ctx context.Context, key T, value any) context.Context {
	return context.WithValue(ctx, key, value)
}
