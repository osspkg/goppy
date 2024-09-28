/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"time"

	"go.osspkg.com/goppy/v2/plugins"
)

// ConfigPgsql pgsql config model
type ConfigPgsql struct {
	Pool    []ConfigPGSqlClient `yaml:"pgsql"`
	Migrate []ConfigMigrateItem `yaml:"pgsql_migrate"`
}

// List getting all configs
func (v *ConfigPgsql) List() (list []ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

func (v *ConfigPgsql) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []ConfigPGSqlClient{
			{
				Tags:        "master",
				Host:        "127.0.0.1",
				Port:        5432,
				Schema:      "test_database",
				User:        "test",
				Password:    "test",
				SSLMode:     false,
				AppName:     "goppy_app",
				MaxIdleConn: 5,
				MaxOpenConn: 5,
				MaxConnTTL:  time.Second * 50,
				Charset:     "UTF8",
				Timeout:     time.Second * 5,
				OtherParams: "",
			},
		}
	}
	if len(v.Migrate) == 0 {
		v.Migrate = []ConfigMigrateItem{
			{
				Tags: "master",
				Dir:  "./migrations",
			},
		}
	}
}

// WithPGSql launch PostgreSQL connection pool
func WithPGSql() plugins.Plugin {
	Register(PgSQLDialect, &_pgsqlMigrateTable{})

	return plugins.Plugin{
		Config: &ConfigPgsql{},
		Inject: func(c *ConfigPgsql, o ORM) error {
			conn := NewPGSqlClient(c)
			o.Register(conn)
			return NewMigrate(o, c.Migrate).
				Run(context.TODO(), conn.Dialect())
		},
	}
}
