/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/geoip2-golang"
	"go.osspkg.com/goppy/plugins"
)

// MaxMindConfig MaxMind database config
type MaxMindConfig struct {
	DB string `yaml:"maxminddb"`
}

func (v *MaxMindConfig) Default() {
	v.DB = "./GeoIP2-City.mmdb"
}

// WithMaxMindGeoIP information resolver through local MaxMind database
func WithMaxMindGeoIP() plugins.Plugin {
	return plugins.Plugin{
		Config: &MaxMindConfig{},
		Inject: func(conf *MaxMindConfig) GeoIP {
			return newMMDB(conf)
		},
	}
}

type (
	//GeoIP geo-ip information definition interface
	GeoIP interface {
		Country(ip net.IP) (string, error)
	}

	maxmind struct {
		conf *MaxMindConfig
		db   *geoip2.Reader
	}
)

func newMMDB(c *MaxMindConfig) *maxmind {
	return &maxmind{
		conf: c,
	}
}

func (v *maxmind) Up() error {
	db, err := geoip2.Open(v.conf.DB)
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
