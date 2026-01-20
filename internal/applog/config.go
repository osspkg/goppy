/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package applog

import "go.osspkg.com/logx"

type (
	GroupConfig struct {
		Log Config `yaml:"log"`
	}

	Config struct {
		Level    uint32 `yaml:"level"`
		FilePath string `yaml:"file_path,omitempty"`
		Format   string `yaml:"format"`
	}
)

func Default() *GroupConfig {
	return &GroupConfig{
		Log: Config{
			Level:    logx.LevelDebug,
			FilePath: "/dev/stdout",
			Format:   "string",
		},
	}
}
