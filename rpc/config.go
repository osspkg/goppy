/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package rpc

import (
	"fmt"
)

type Type string

const (
	TypeGoPlugin   Type = "goplugin"
	TypeUnixSocket Type = "unix"
)

type ConfigGroup struct {
	Items []Config `yaml:"rpc"`
}

type Config struct {
	Name    string            `yaml:"name"`
	Type    Type              `yaml:"type"`
	Path    string            `yaml:"path"`
	Options map[string]string `json:"options,omitempty"`
}

func (v *ConfigGroup) Default() {
	if len(v.Items) == 0 {
		v.Items = append(v.Items,
			Config{
				Name:    "go-plugin",
				Type:    TypeGoPlugin,
				Path:    "plugin.so",
				Options: nil,
			},
			Config{
				Name: "unix-plugin",
				Type: TypeUnixSocket,
				Path: "plugin.bin",
				Options: map[string]string{
					"proto": "jsonrpc",
					"pid":   "/tmp/plugin.pid",
				},
			},
		)
	}
}

func (v *ConfigGroup) Validate() error {
	if len(v.Items) == 0 {
		return fmt.Errorf("rpc: empty config")
	}

	return nil
}
