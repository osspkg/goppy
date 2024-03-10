package geoip

import (
	"context"
	"net"
	"strings"
)

const (
	geoClientIP    = "x-geo-client-ip"
	geoProxyIPs    = "X-geo-proxy-ips"
	geoCountryName = "x-geo-country"
)

type geoIPContext string

func SetClientIP(ctx context.Context, value net.IP) context.Context {
	return context.WithValue(ctx, geoIPContext(geoClientIP), value)
}

func GetClientIP(ctx context.Context) net.IP {
	value, ok := ctx.Value(geoIPContext(geoClientIP)).(net.IP)
	if !ok {
		return nil
	}
	return value
}

func SetProxyIPs(ctx context.Context, value []net.IP) context.Context {
	return context.WithValue(ctx, geoIPContext(geoProxyIPs), value)
}

func GetProxyIPs(ctx context.Context) []net.IP {
	value, ok := ctx.Value(geoIPContext(geoProxyIPs)).([]net.IP)
	if !ok {
		return nil
	}
	return value
}

func SetCountryName(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, geoIPContext(geoCountryName), value)
}

func GetCountryName(ctx context.Context) *string {
	value, ok := ctx.Value(geoIPContext(geoCountryName)).(string)
	if !ok || value == "XX" || value == "" {
		return nil
	}
	return &value
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
