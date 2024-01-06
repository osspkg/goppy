/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.osspkg.com/goppy/auth/jwt"
	"go.osspkg.com/goppy/web"
)

const (
	jwtPayload = "jwtp"
	jwtHeader  = "jwth"
)

type (
	jwtContext string

	ctx interface {
		Context() context.Context
	}

	JWTGuardMiddlewareConfig struct {
		AcceptHeader bool
		AcceptCookie string
	}
)

func setJWTPayloadContext(ctx context.Context, value []byte) context.Context {
	return context.WithValue(ctx, jwtContext(jwtPayload), value)
}

func GetJWTPayloadContext(c ctx, payload interface{}) error {
	value, ok := c.Context().Value(jwtContext(jwtPayload)).([]byte)
	if !ok {
		return fmt.Errorf("jwt payload not found")
	}
	return json.Unmarshal(value, payload)
}

func setJWTHeaderContext(ctx context.Context, value *jwt.Header) context.Context {
	return context.WithValue(ctx, jwtContext(jwtHeader), *value)
}

func GetJWTHeaderContext(c ctx) *jwt.Header {
	value, ok := c.Context().Value(jwtContext(jwtPayload)).(jwt.Header)
	if !ok {
		return nil
	}
	return &value
}

func JWTGuardMiddleware(j JWT, c JWTGuardMiddlewareConfig) web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			val := ""

			if c.AcceptHeader {
				val = r.Header.Get("Authorization")
				if len(val) > 7 && strings.HasPrefix(val, "Bearer ") {
					val = val[6:]
				}
			}

			if len(val) == 0 && len(c.AcceptCookie) > 0 {
				cv, err := r.Cookie(c.AcceptCookie)
				if err == nil && len(cv.Value) > 0 {
					val = cv.Value
				}
			}

			if len(val) == 0 {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			var raw json.RawMessage
			h, err := j.Verify(val, &raw)
			if err != nil {
				http.Error(w, "authorization required", http.StatusUnauthorized)
				return
			}

			ctx = setJWTHeaderContext(ctx, h)
			ctx = setJWTPayloadContext(ctx, raw)

			call(w, r.WithContext(ctx))
		}
	}
}
