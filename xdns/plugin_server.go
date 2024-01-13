/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

func WithDNSServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l xlog.Logger) *Server {
			return NewServer(c.DNS, l)
		},
	}
}
