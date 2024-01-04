/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

// ConfigDebug config to initialize HTTP debug service
type ConfigDebug struct {
	Config ConfigHttp `yaml:"debug"`
}

func (v *ConfigDebug) Default() {
	v.Config = ConfigHttp{Addr: "127.0.0.1:12000"}
}

// WithHTTPDebug debug service over HTTP protocol with pprof enabled
func WithHTTPDebug() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigDebug{},
		Inject: func(c *ConfigDebug, l xlog.Logger) Debugger {
			return NewServerDebug(c.Config, l)
		},
	}
}

type Debugger interface {
}
