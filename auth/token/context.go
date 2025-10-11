/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"context"
	"encoding/json"
)

type ctxValue string

const (
	tokenPayload ctxValue = "jwt_payload"
	tokenHeader  ctxValue = "jwt_header"
)

func SetPayloadContext[T json.Unmarshaler](ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, tokenPayload, value)
}

func GetPayloadContext[T json.Unmarshaler](ctx context.Context) (T, bool) {
	value, ok := ctx.Value(tokenPayload).(T)
	return value, ok
}

func SetHeaderContext(ctx context.Context, value Header) context.Context {
	return context.WithValue(ctx, tokenHeader, value)
}

func GetHeaderContext(ctx context.Context) (Header, bool) {
	value, ok := ctx.Value(tokenHeader).(Header)
	return value, ok
}
