/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v2/plugins"
)

func WithORM() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(ctx xc.Context) ORM {
			o := New(ctx.Context())
			go func() {
				select {
				case <-ctx.Done():
					o.Close()
				}
			}()
			return o
		},
	}
}
