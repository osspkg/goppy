/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"go.osspkg.com/goppy/v3/plugin"
	"go.osspkg.com/goppy/v3/plugins/web"
)

func WithTransport(opts ...Option) plugin.Kind {
	return plugin.Kind{
		Inject: func(r web.ServerPool) Transport {
			return newService(r, opts...)
		},
	}
}
