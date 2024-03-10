/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xlog"
)

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigDNS{},
		Inject: func(c *ConfigDNS, l xlog.Logger) *Server {
			return NewServer(c.DNS, l)
		},
	}
}
