/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"net"
	"net/http"

	"go.osspkg.com/goppy/v2/web"
)

// CloudflareMiddleware determine geo-ip information when proxying through Cloudflare
func CloudflareMiddleware() web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cip := net.ParseIP(r.Header.Get("CF-Connecting-IP"))
			ctx = SetCountryName(ctx, r.Header.Get("CF-IPCountry"))
			ctx = SetClientIP(ctx, cip)
			ctx = SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))
			call(w, r.WithContext(ctx))
		}
	}
}
