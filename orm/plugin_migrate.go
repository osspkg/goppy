/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v2/plugins"
)

func WithMigration(ms ...Migration) plugins.Plugin {
	if len(ms) > 0 {
		return plugins.Plugin{
			Inject: func(ctx xc.Context, o ORM) error {
				m := &Migrate{
					Conn: o,
					FS:   newMemFS(ms),
				}
				go dialectOnRegistered(func(dialect string) {
					if err := m.Run(ctx.Context(), dialect); err != nil {
						logx.Error("Run DB migration", "dialect", dialect, "err", err)
						ctx.Close()
						return
					}
				})
				return nil
			},
		}
	}
	return plugins.Plugin{
		Config: &ConfigMigrate{},
		Inject: func(ctx xc.Context, c *ConfigMigrate, o ORM) error {
			m := &Migrate{
				Conn: o,
				FS:   newOSFS(c.List),
			}
			go dialectOnRegistered(func(dialect string) {
				if err := m.Run(ctx.Context(), dialect); err != nil {
					logx.Error("Run DB migration", "dialect", dialect, "err", err)
					ctx.Close()
					return
				}
			})
			return nil
		},
	}
}
