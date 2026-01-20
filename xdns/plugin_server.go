/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import (
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
)

func WithServer() plugins.Kind {
	return plugins.Kind{
		Config: &ConfigGroup{},
		Inject: func(ctx xc.Context, c *ConfigGroup) *Server {
			return NewServer(ctx.Context(), c.DNS)
		},
	}
}
