/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"go.osspkg.com/goppy/v2/plugins"
)

// ConfigGroup config to initialize HTTP service
type ConfigGroup struct {
	HTTP []Config `yaml:"http"`
}

func (v *ConfigGroup) Default() {
	if v.HTTP == nil {
		v.HTTP = append(v.HTTP, Config{
			Tag:  "main",
			Addr: "0.0.0.0:8080",
		})
	}
}

// WithServer launch of HTTP service with default Router
func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigGroup{},
		Inject: func(conf *ConfigGroup) RouterPool {
			return newRouteProvider(conf.HTTP)
		},
	}
}
