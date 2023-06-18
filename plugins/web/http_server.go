/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"github.com/osspkg/go-sdk/log"
	"github.com/osspkg/go-sdk/webutil"
	"github.com/osspkg/goppy/plugins"
)

// ConfigHttp config to initialize HTTP service
type ConfigHttp struct {
	Config map[string]webutil.ConfigHttp `yaml:"http"`
}

func (v *ConfigHttp) Default() {
	if v.Config == nil {
		v.Config = map[string]webutil.ConfigHttp{
			"main": {Addr: "127.0.0.1:8080"},
		}
	}
}

// WithHTTP launch of HTTP service with default Router
func WithHTTP() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigHttp{},
		Inject: func(conf *ConfigHttp, l log.Logger) (*routeProvider, RouterPool) {
			rp := newRouteProvider(conf.Config, l)
			return rp, rp
		},
	}
}
