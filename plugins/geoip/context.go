package geoip

import (
	"context"
	"net"
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
