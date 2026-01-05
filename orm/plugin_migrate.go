/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"go.osspkg.com/logx"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
)

func WithMigration(ms ...Migration) plugins.Kind {
	if len(ms) > 0 {
		return plugins.Kind{
			Inject: func(ctx xc.Context, o ORM) error {
				if err := NewMigrate(o, NewVirtualFS(ms)).Run(ctx.Context()); err != nil {
					logx.Error("Run DB migration", "err", err)
					return err
				}

				return nil
			},
		}
	}

	return plugins.Kind{
		Config: &ConfigGroup{},
		Inject: func(ctx xc.Context, c *ConfigGroup, o ORM) error {
			if err := NewMigrate(o, NewOperationSystemFS(c.List)).Run(ctx.Context()); err != nil {
				logx.Error("Run DB migration", "err", err)
				return err
			}

			return nil
		},
	}
}
