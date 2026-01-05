/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import "context"

type webCtx string

func SetContext(ctx context.Context, key string, value any) context.Context {
	return context.WithValue(ctx, webCtx(key), value)
}

func GetContext[T any](ctx context.Context, key string) (T, bool) {
	val, ok := ctx.Value(webCtx(key)).(T)
	return val, ok
}
