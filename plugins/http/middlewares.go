package http

import (
	"bytes"
	"net"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/dewep-online/goppy/plugins/geoip"
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

// CloudflareMiddleware determine geo-ip information when proxying through Cloudflare
func CloudflareMiddleware() Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = geoip.SetCountryName(ctx, r.Header.Get("CF-IPCountry"))

			cip := net.ParseIP(r.Header.Get("CF-Connecting-IP"))
			if len(cip) == 0 {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					cip = net.ParseIP(host)
				}
			}
			ctx = geoip.SetClientIP(ctx, cip)

			ctx = geoip.SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))

			call(w, r.WithContext(ctx))
		}
	}
}

type geoIP interface {
	Country(ip net.IP) (string, error)
}

// MaxMindMiddleware determine geo-ip information through local MaxMind database
func MaxMindMiddleware(resolver geoIP) Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			cip := geoip.GetClientIP(r)
			if len(cip) == 0 {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					cip = net.ParseIP(host)
				}
				ctx = geoip.SetClientIP(ctx, cip)
			}

			country, _ := resolver.Country(cip) //nolint: errcheck
			ctx = geoip.SetCountryName(ctx, country)

			if ips := geoip.GetProxyIPs(r); len(ips) == 0 {
				ctx = geoip.SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))
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
		if !bytes.Equal(skip, ip) && ip != nil {
			result = append(result, ip)
		}
	}
	return result
}
