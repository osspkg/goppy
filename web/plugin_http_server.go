/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

// ConfigHttpPool config to initialize HTTP service
type ConfigHttpPool struct {
	Config map[string]Config `yaml:"http"`
}

func (v *ConfigHttpPool) Default() {
	if v.Config == nil {
		v.Config = map[string]Config{
			"main": {Addr: "0.0.0.0:8080"},
		}
	}
}

// WithServer launch of HTTP service with default Router
func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigHttpPool{},
		Inject: func(conf *ConfigHttpPool, l xlog.Logger) RouterPool {
			return newRouteProvider(conf.Config, l)
		},
	}
}
