/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import "fmt"

const (
	customAnalyzerName = "goppy_analyzer"
	customFilterName   = "goppy_filter"
)

type (
	ConfigGroup struct {
		Search Config `yaml:"search"`
	}
	Config struct {
		Folder string `yaml:"folder"`
	}
)

func (c Config) Validate() error {
	if len(c.Folder) == 0 {
		return fmt.Errorf("storage folder is required")
	}
	return nil
}
