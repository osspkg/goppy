/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package geoip

type ConfigGroup struct {
	GeoIP Config `yaml:"geoip"`
}

type Config struct {
	MaxMindDB string `yaml:"maxminddb"`
}

func (v *ConfigGroup) Default() {
	v.GeoIP.MaxMindDB = "./GeoIP2-City.mmdb"
}
