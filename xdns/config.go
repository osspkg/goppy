/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import "time"

type (
	ConfigGroup struct {
		DNS Config `yaml:"dns"`
	}
	Config struct {
		Addr    string        `yaml:"addr"`
		Timeout time.Duration `yaml:"timeout"`
	}
)

func (v *ConfigGroup) Default() {
	v.DNS.Addr = "0.0.0.0:53"
	v.DNS.Timeout = 5 * time.Second
}
