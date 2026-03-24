/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"go.osspkg.com/goppy/v3/plugins"
	"go.osspkg.com/goppy/v3/web"
)

func WithTransport(opts ...Option) plugins.Kind {
	return plugins.Kind{
		Inject: func(r web.ServerPool) Transport {
			return newTransport(r, opts...)
		},
	}
}
