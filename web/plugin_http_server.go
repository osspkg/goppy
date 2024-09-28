/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"go.osspkg.com/goppy/v2/plugins"
)

// ConfigHttpPool config to initialize HTTP service
type ConfigHttpPool struct {
	Config []Config `yaml:"http"`
}

func (v *ConfigHttpPool) Default() {
	if v.Config == nil {
		v.Config = append(v.Config, Config{
			Tag:  "main",
			Addr: "0.0.0.0:8080",
		})
	}
}

// WithServer launch of HTTP service with default Router
func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigHttpPool{},
		Inject: func(conf *ConfigHttpPool) RouterPool {
			return newRouteProvider(conf.Config)
		},
	}
}
