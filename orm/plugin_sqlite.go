/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"go.osspkg.com/goppy/v2/plugins"
)

// ConfigSqlite sqlite config model
type ConfigSqlite struct {
	Pool []ConfigSqliteClient `yaml:"sqlite"`
}

func (v *ConfigSqlite) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []ConfigSqliteClient{
			{
				Tags:        "master",
				File:        "./sqlite.db",
				Cache:       "private",
				Mode:        "rwc",
				Journal:     "WAL",
				LockingMode: "EXCLUSIVE",
				OtherParams: "auto_vacuum=incremental",
			},
		}
	}
}

// List getting all configs
func (v *ConfigSqlite) List() (list []ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

// WithSqliteClient launch SQLite connection pool
func WithSqliteClient() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigSqlite{},
		Inject: func(c *ConfigSqlite, o ORM) error {
			conn := NewSqliteClient(c)
			return o.Register(conn, func() {
				dialectRegister(SQLiteDialect, &_sqliteMigrateTable{})
			})
		},
	}
}
