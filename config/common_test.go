/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config_test

import (
	"os"
	"testing"

	"go.osspkg.com/goppy/config"
	"go.osspkg.com/goppy/xtest"
)

type (
	testConfigItem struct {
		Home string `yaml:"home"`
		Path string `yaml:"path"`
	}
	testConfig struct {
		Envs testConfigItem `yaml:"envs"`
	}
)

func TestUnit_ConfigResolve(t *testing.T) {
	filename := "/tmp/TestUnit_ConfigResolve.yaml"
	data := `
envs:
  home: "@env(HOME#fail)"
  path: "@env(PATH#fail)"
`
	err := os.WriteFile(filename, []byte(data), 0755)
	xtest.NoError(t, err)

	res := config.New(config.EnvResolver())

	err = res.OpenFile(filename)
	xtest.NoError(t, err)
	err = res.Build()
	xtest.NoError(t, err)

	tc := &testConfig{}

	err = res.Decode(tc)
	xtest.NoError(t, err)
	xtest.NotEqual(t, "fail", tc.Envs.Home)
	xtest.NotEqual(t, "fail", tc.Envs.Path)
}
