/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"context"
	"net"
	"strings"

	"go.osspkg.com/do"
)

type geoIPContext string

const (
	geoClientIP    geoIPContext = "x-geo-client-ip"
	geoProxyIPs    geoIPContext = "X-geo-proxy-ips"
	geoCountryName geoIPContext = "x-geo-country"
)

func SetClientIP(ctx context.Context, value net.IP) context.Context {
	if len(value) == 0 {
		return ctx
	}
	return context.WithValue(ctx, geoClientIP, value)
}

func GetClientIP(ctx context.Context) net.IP {
	value, ok := ctx.Value(geoClientIP).(net.IP)
	if !ok {
		return net.IPv4zero
	}
	return value
}

func SetProxyIPs(ctx context.Context, value []net.IP) context.Context {
	if len(value) == 0 {
		return ctx
	}
	return context.WithValue(ctx, geoProxyIPs, value)
}

func GetProxyIPs(ctx context.Context) []net.IP {
	value, ok := ctx.Value(geoProxyIPs).([]net.IP)
	if !ok {
		return nil
	}
	return value
}

func SetCountryName(ctx context.Context, value string) context.Context {
	if len(value) == 0 {
		return ctx
	}
	return context.WithValue(ctx, geoCountryName, value)
}

func GetCountryName(ctx context.Context) string {
	value, ok := ctx.Value(geoCountryName).(string)
	if !ok {
		return UnknownCountry
	}
	return value
}

func ParseIPs(values string, skip ...string) []net.IP {
	var result []net.IP

	skipMap := do.ToMap(skip)

	for _, v := range strings.Split(values, ",") {
		if len(v) == 0 {
			continue
		}

		v = strings.TrimSpace(v)
		if _, ok := skipMap[v]; ok {
			continue
		}

		if ip := net.ParseIP(v); ip != nil {
			result = append(result, ip)
		}
	}

	return result
}
