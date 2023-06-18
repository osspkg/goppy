/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/osspkg/goppy/plugins/web"
)

const (
	geoClientIP    = "x-geo-client-ip"
	geoProxyIPs    = "X-geo-proxy-ips"
	geoCountryName = "x-geo-country"
)

type geoIPContext string

type ctx interface {
	Context() context.Context
}

func SetClientIP(ctx context.Context, value net.IP) context.Context {
	return context.WithValue(ctx, geoIPContext(geoClientIP), value)
}

func GetClientIP(c ctx) net.IP {
	value, ok := c.Context().Value(geoIPContext(geoClientIP)).(net.IP)
	if !ok {
		return nil
	}
	return value
}

func SetProxyIPs(ctx context.Context, value []net.IP) context.Context {
	return context.WithValue(ctx, geoIPContext(geoProxyIPs), value)
}

func GetProxyIPs(c ctx) []net.IP {
	value, ok := c.Context().Value(geoIPContext(geoProxyIPs)).([]net.IP)
	if !ok {
		return nil
	}
	return value
}

func SetCountryName(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, geoIPContext(geoCountryName), value)
}

func GetCountryName(c ctx) *string {
	value, ok := c.Context().Value(geoIPContext(geoCountryName)).(string)
	if !ok || value == "XX" || value == "" {
		return nil
	}
	return &value
}

// CloudflareMiddleware determine geo-ip information when proxying through Cloudflare
func CloudflareMiddleware() web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = SetCountryName(ctx, r.Header.Get("CF-IPCountry"))
			cip := net.ParseIP(r.Header.Get("CF-Connecting-IP"))
			if len(cip) == 0 {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					cip = net.ParseIP(host)
				}
			}
			ctx = SetClientIP(ctx, cip)
			ctx = SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))
			call(w, r.WithContext(ctx))
		}
	}
}

type geoIP interface {
	Country(ip net.IP) (string, error)
}

// MaxMindMiddleware determine geo-ip information through local MaxMind database
func MaxMindMiddleware(resolver geoIP) web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cip := GetClientIP(r)
			if len(cip) == 0 {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					cip = net.ParseIP(host)
				}
				ctx = SetClientIP(ctx, cip)
			}
			country, _ := resolver.Country(cip) //nolint: errcheck
			ctx = SetCountryName(ctx, country)
			if ips := GetProxyIPs(r); len(ips) == 0 {
				ctx = SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))
			}
			call(w, r.WithContext(ctx))
		}
	}
}

func parseXForwardedFor(ff string, skip net.IP) []net.IP {
	var result []net.IP
	for _, v := range strings.Split(ff, ",") {
		if len(v) == 0 {
			continue
		}
		ip := net.ParseIP(strings.TrimSpace(v))
		if !skip.Equal(ip) && ip != nil {
			result = append(result, ip)
		}
	}
	return result
}
