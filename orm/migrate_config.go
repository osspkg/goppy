/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import "go.osspkg.com/goppy/v3/orm/dialect"

type (
	ConfigGroup struct {
		List []Config `yaml:"db_migrate"`
	}

	Config struct {
		Tags    string       `yaml:"tags"`
		Dialect dialect.Name `yaml:"dialect"`
		Dir     string       `yaml:"dir"`
	}
)

func (v *ConfigGroup) Default() {
	if len(v.List) == 0 {
		v.List = []Config{
			{
				Tags:    "master",
				Dialect: "unknown",
				Dir:     "./migrations/unknown",
			},
		}
	}
}
