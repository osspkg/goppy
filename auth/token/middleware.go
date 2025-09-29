/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.osspkg.com/goppy/v2/web"
)

func GuardMiddleware[T json.Unmarshaler](srv Token) web.Middleware {
	return GuardMiddlewareCustom[T](srv, nil, nil, nil)
}

func GuardMiddlewareCustom[T json.Unmarshaler](
	srv Token,
	before, after func(web.Ctx) (broke bool),
	fail func(web.Ctx, error),
) web.Middleware {
	return func(call func(web.Ctx)) func(web.Ctx) {
		return func(wc web.Ctx) {
			if before != nil && before(wc) {
				return
			}

			val := ""

			if key := srv.HeaderName(); len(key) > 0 {
				val = wc.Header().Get(key)
				if len(val) > 7 && strings.HasPrefix(val, "Bearer ") {
					val = val[6:]
				}
			}

			if key := srv.CookieName(); len(val) == 0 && len(key) > 0 {
				val = wc.Cookie().Get(key)
			}

			if len(val) == 0 {
				if fail != nil {
					fail(wc, ErrEmptyToken)
				} else {
					wc.Error(http.StatusUnauthorized, ErrEmptyToken)
				}
				return
			}

			head, payload, err := srv.VerifyJWT([]byte(val))
			if err != nil {
				if fail != nil {
					fail(wc, fmt.Errorf("failed to verify token: %w", err))
				} else {
					wc.Error(http.StatusUnauthorized, ErrInvalidToken)
				}
				return
			}

			var data T
			if err = json.Unmarshal(payload, &data); err != nil {
				if fail != nil {
					fail(wc, fmt.Errorf("failed unmarshal token payload: %w", err))
				} else {
					wc.Error(http.StatusUnauthorized, ErrInvalidToken)
				}
				return
			}

			wc.SetContextValue(tokenHeader, *head)
			wc.SetContextValue(tokenPayload, data)

			if after != nil && after(wc) {
				return
			}

			call(wc)
		}
	}
}
