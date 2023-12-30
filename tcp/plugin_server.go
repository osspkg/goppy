/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"time"

	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/tcp/server"
	"go.osspkg.com/goppy/xlog"
)

type Config struct {
	TCP server.ConfigItem `yaml:"tcp"`
}

func (v *Config) Default() {
	if len(v.TCP.Pools) == 0 {
		v.TCP.Pools = append(v.TCP.Pools, server.Pool{
			Port: 8080,
			Certs: []server.Cert{{
				Public:  "./ssl/public.crt",
				Private: "./ssl/private.key",
			}},
		})
	}
	if v.TCP.Timeout == 0 {
		v.TCP.Timeout = 10 * time.Second
	}
}

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l xlog.Logger) *server.Server {
			return server.New(c.TCP, l)
		},
	}
}
