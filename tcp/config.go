/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import "time"

type ConfigItem struct {
	Address           string        `yaml:"address"`
	Certs             []Cert        `yaml:"certs,omitempty"`
	Timeout           time.Duration `yaml:"timeout,omitempty"`
	ClientMaxBodySize int           `yaml:"client_max_body_size,omitempty"`
}

type Cert struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}
