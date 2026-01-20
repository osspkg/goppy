/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import "fmt"

type (
	ConfigGroup struct {
		DNS Config `yaml:"dns"`
	}
	Config struct {
		Addr   string   `yaml:"addr"`
		QTypes []string `yaml:"qtypes,omitempty"`
	}
)

func (v *ConfigGroup) Default() {
	v.DNS.Addr = "0.0.0.0:53"
	v.DNS.QTypes = []string{"A"}
}

func (v *ConfigGroup) Validate() error {
	if len(v.DNS.Addr) == 0 {
		return fmt.Errorf("dns server: missing addr")
	}
	return nil
}
