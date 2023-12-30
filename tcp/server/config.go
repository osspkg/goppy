/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import "time"

type ConfigItem struct {
	Pools   []Pool        `yaml:"pool"`
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

type Pool struct {
	Port  int    `yaml:"port"`
	Certs []Cert `yaml:"certs,omitempty"`
}

type Cert struct {
	Public  string `yaml:"pub"`
	Private string `yaml:"priv"`
}
