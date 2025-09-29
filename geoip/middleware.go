/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"net"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v2/web"
)

func ResolveIPMiddleware(resolver GeoIP, headers ...string) web.Middleware {
	if len(headers) == 0 {
		headers = append(headers, HeaderXRealIP)
	}
	return func(call func(web.Ctx)) func(web.Ctx) {
		return func(ctx web.Ctx) {

			for _, header := range headers {
				if val := ctx.Header().Get(header); len(val) > 0 {
					cip := net.ParseIP(val)
					ctx.SetContextValue(geoClientIP, cip)
					break
				}
			}

			if val := ctx.Header().Get(HeaderXForwardedFor); len(val) > 0 {
				cip := GetClientIP(ctx.Context())
				ctx.SetContextValue(geoProxyIPs, ParseIPs(val, cip.String()))
			}

			if resolver != nil {
				cip := GetClientIP(ctx.Context())
				if country, err := resolver.Country(cip); err != nil {
					logx.Warn("GeoIP country lookup failed",
						"err", err,
						"ip", cip.String(),
						"forwarded", ctx.Header().Get(HeaderXForwardedFor),
					)
				} else {
					ctx.SetContextValue(geoCountryName, country)
				}
			}

			call(ctx)
		}
	}
}

func HeadersMiddleware(ipHeader, countryHeader string) web.Middleware {
	return func(call func(web.Ctx)) func(web.Ctx) {
		return func(ctx web.Ctx) {

			if val := ctx.Header().Get(ipHeader); len(val) > 0 {
				cip := net.ParseIP(val)
				ctx.SetContextValue(geoClientIP, cip)
			}

			if val := ctx.Header().Get(countryHeader); len(val) > 0 {
				ctx.SetContextValue(geoCountryName, val)
			}

			if val := ctx.Header().Get(HeaderXForwardedFor); len(val) > 0 {
				cip := GetClientIP(ctx.Context())
				ctx.SetContextValue(geoProxyIPs, ParseIPs(val, cip.String()))
			}

			call(ctx)
		}
	}
}
