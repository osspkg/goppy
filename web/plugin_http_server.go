/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

// ConfigHttpPool config to initialize HTTP service
type ConfigHttpPool struct {
	Config map[string]ConfigHttp `yaml:"http"`
}

func (v *ConfigHttpPool) Default() {
	if v.Config == nil {
		v.Config = map[string]ConfigHttp{
			"main": {Addr: "127.0.0.1:8080"},
		}
	}
}

// WithHTTP launch of HTTP service with default Router
func WithHTTP() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigHttpPool{},
		Inject: func(conf *ConfigHttpPool, l xlog.Logger) (*routeProvider, RouterPool) {
			rp := newRouteProvider(conf.Config, l)
			return rp, rp
		},
	}
}
