/*
 *  Copyright (c) 2024-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config_test

import (
	"os"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/pkg/config"
)

func TestUnit_ConfigResolve(t *testing.T) {
	type (
		TestConfigItem struct {
			Home string `yaml:"home"`
			Path string `yaml:"path"`
			Tmp  string `yaml:"tempenv"`
		}
		TestConfig struct {
			Envs TestConfigItem `yaml:"envs"`
		}
	)

	data := `
envs:
    home: "@env(HOME#fail)"
    path: "@env(PATH#fail)"
`

	casecheck.NoError(t, os.Setenv("tempenv", "123"))

	res := config.New(config.NewEnvResolver())
	res.OpenBlob(data, ".yaml")
	casecheck.NoError(t, res.Build())

	var tc TestConfig

	casecheck.NoError(t, res.Decode(&tc))
	casecheck.NotEqual(t, "fail", tc.Envs.Home)
	casecheck.NotEqual(t, "fail", tc.Envs.Path)
	casecheck.NotEqual(t, "123", tc.Envs.Tmp)

	res.Flush()
	casecheck.Error(t, res.Build())
	casecheck.Error(t, res.Decode(&tc))
}
