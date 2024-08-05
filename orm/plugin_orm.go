/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"
)

func WithORM() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(ctx xc.Context, l logx.Logger) ORM {
			o := New(ctx.Context(), l)
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
