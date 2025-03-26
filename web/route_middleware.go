/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"
	"sync/atomic"

	"go.osspkg.com/logx"
)

// Middleware type of middleware
type Middleware func(func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request)

// ThrottlingMiddleware limits active requests
func ThrottlingMiddleware(max int64) Middleware {
	var i int64
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if atomic.LoadInt64(&i) >= max {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			atomic.AddInt64(&i, 1)
			call(w, r)
			atomic.AddInt64(&i, -1)
		}
	}
}

// RecoveryMiddleware recovery go panic and write to log
func RecoveryMiddleware() func(
	func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logx.Error("Panic recovered", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			f(w, r)
		}
	}
}

func HeadersContextWrapMiddleware(args ...string) func(
	func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, arg := range args {
				val := r.Header.Get(arg)
				if len(val) == 0 {
					continue
				}
				r = r.WithContext(SetContext(r.Context(), arg, val))
			}

			call(w, r)
		}
	}
}

func CookiesContextWrapMiddleware(args ...string) func(
	func(http.ResponseWriter, *http.Request),
) func(http.ResponseWriter, *http.Request) {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			for _, arg := range args {
				val, err := r.Cookie(arg)
				if err != nil {
					continue
				}
				r = r.WithContext(SetContext(r.Context(), arg, val.Value))
			}

			call(w, r)
		}
	}
}
