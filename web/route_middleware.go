/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"
	"sync/atomic"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
)

// Middleware type of middleware
type Middleware func(func(Ctx)) func(Ctx)

// ThrottlingMiddleware limits active requests
func ThrottlingMiddleware(max int64) Middleware {
	var i int64
	err := errors.New(http.StatusText(http.StatusTooManyRequests))

	return func(call func(Ctx)) func(Ctx) {
		return func(ctx Ctx) {
			if atomic.LoadInt64(&i) >= max {
				ctx.Error(http.StatusTooManyRequests, err)
				return
			}

			atomic.AddInt64(&i, 1)
			call(ctx)
			atomic.AddInt64(&i, -1)
		}
	}
}

// RecoveryMiddleware recovery go panic and write to log
func RecoveryMiddleware() Middleware {
	return func(call func(Ctx)) func(Ctx) {
		err := errors.New(http.StatusText(http.StatusInternalServerError))

		return func(ctx Ctx) {
			defer func() {
				if e := recover(); e != nil {
					logx.Error("web.RecoveryMiddleware", "err", e)
					ctx.Error(http.StatusInternalServerError, err)
				}
			}()
			call(ctx)
		}
	}
}

func HeadersContextWrapMiddleware(args ...string) Middleware {
	return func(call func(Ctx)) func(Ctx) {
		return func(ctx Ctx) {
			headers := ctx.Header()

			for _, arg := range args {
				val := headers.Get(arg)
				if len(val) == 0 {
					continue
				}

				ctx.SetContextValue(webCtx(arg), val)
			}

			call(ctx)
		}
	}
}

func CookiesContextWrapMiddleware(args ...string) Middleware {
	return func(call func(Ctx)) func(Ctx) {
		return func(ctx Ctx) {
			cookies := ctx.Cookie()

			for _, arg := range args {
				val := cookies.Get(arg)
				if len(val) == 0 {
					continue
				}

				ctx.SetContextValue(webCtx(arg), val)
			}

			call(ctx)
		}
	}
}
