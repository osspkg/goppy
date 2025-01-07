/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"fmt"
	"net"
	"net/http"

	"github.com/oschwald/geoip2-golang"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
)

type ConfigMaxMind struct {
	GeoIP struct {
		MaxMindDB string `yaml:"maxminddb"`
	} `yaml:"geoip"`
}

func (v *ConfigMaxMind) Default() {
	v.GeoIP.MaxMindDB = "./GeoIP2-City.mmdb"
}

// WithMaxMindGeoIP information resolver through local MaxMind database
func WithMaxMindGeoIP() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigMaxMind{},
		Inject: func(conf *ConfigMaxMind) GeoIP {
			return newMaxMindGeoIP(conf)
		},
	}
}

type (
	// GeoIP geo-ip information definition interface
	GeoIP interface {
		Country(ip net.IP) (string, error)
	}

	maxmind struct {
		conf *ConfigMaxMind
		db   *geoip2.Reader
	}
)

func newMaxMindGeoIP(c *ConfigMaxMind) *maxmind {
	return &maxmind{
		conf: c,
	}
}

func (v *maxmind) Up() error {
	db, err := geoip2.Open(v.conf.GeoIP.MaxMindDB)
	if err != nil {
		return fmt.Errorf("maxmind: %w", err)
	}
	v.db = db
	return nil
}

func (v *maxmind) Down() error {
	if v.db != nil {
		return v.db.Close()
	}
	return nil
}

func (v *maxmind) Country(ip net.IP) (string, error) {
	vv, err := v.db.Country(ip)
	if err != nil {
		return "", err
	}
	return vv.Country.IsoCode, nil
}

// MaxMindMiddleware determine geo-ip information through local MaxMind database
func MaxMindMiddleware(resolver GeoIP) web.Middleware {
	return func(call func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			cip := net.ParseIP(r.Header.Get("X-Real-IP"))
			country, _ := resolver.Country(cip) // nolint: errcheck
			ctx = SetCountryName(ctx, country)
			ctx = SetClientIP(ctx, cip)
			ctx = SetProxyIPs(ctx, parseXForwardedFor(r.Header.Get("X-Forwarded-For"), cip))
			call(w, r.WithContext(ctx))
		}
	}
}
