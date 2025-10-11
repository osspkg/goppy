/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"

	"go.osspkg.com/goppy/v2/plugins"
)

// WithMaxMindGeoIP information resolver through local MaxMind database
func WithMaxMindGeoIP() plugins.Kind {
	return plugins.Kind{
		Config: &ConfigGroup{},
		Inject: func(conf *ConfigGroup) GeoIP {
			return NewMaxMindGeoIP(conf)
		},
	}
}

type (
	// GeoIP geo-ip information definition interface
	GeoIP interface {
		Country(ip net.IP) (string, error)
	}

	maxmind struct {
		conf *ConfigGroup
		db   *geoip2.Reader
	}
)

func NewMaxMindGeoIP(c *ConfigGroup) GeoIP {
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
	value, err := v.db.Country(ip)
	if err != nil {
		return "", err
	}
	return value.Country.IsoCode, nil
}
