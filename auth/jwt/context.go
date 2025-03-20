/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jwt

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	jwtPayload jwtContextValue = "jwt_payload"
)

type (
	jwtContextValue string
)

func setJWTPayloadContext(ctx context.Context, value []byte) context.Context {
	return context.WithValue(ctx, jwtPayload, value)
}

func PayloadContext(ctx context.Context, payload any) error {
	value, ok := ctx.Value(jwtPayload).([]byte)
	if !ok {
		return fmt.Errorf("jwt payload not found")
	}
	return json.Unmarshal(value, payload)
}
