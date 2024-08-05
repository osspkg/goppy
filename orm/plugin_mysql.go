/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"time"

	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/logx"
)

// ConfigMysql mysql config model
type ConfigMysql struct {
	Pool    []ConfigMysqlClient `yaml:"mysql"`
	Migrate []ConfigMigrateItem `yaml:"mysql_migrate"`
}

// List getting all configs
func (v *ConfigMysql) List() (list []ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

func (v *ConfigMysql) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []ConfigMysqlClient{
			{
				Tags:              "master",
				Host:              "127.0.0.1",
				Port:              3306,
				Schema:            "test_database",
				User:              "test",
				Password:          "test",
				MaxIdleConn:       5,
				MaxOpenConn:       5,
				MaxConnTTL:        time.Second * 50,
				InterpolateParams: false,
				Timezone:          "UTC",
				TxIsolationLevel:  "",
				Charset:           "utf8mb4",
				Collation:         "utf8mb4_unicode_ci",
				Timeout:           time.Second * 5,
				ReadTimeout:       time.Second * 5,
				WriteTimeout:      time.Second * 5,
				OtherParams:       "",
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

// WithMysql launch MySQL connection pool
func WithMysql() plugins.Plugin {
	Register(MySQLDialect, &_mysqlMigrateTable{})

	return plugins.Plugin{
		Config: &ConfigMysql{},
		Inject: func(c *ConfigMysql, o ORM, l logx.Logger) error {
			conn := NewMysqlClient(c)
			o.Register(conn)
			return NewMigrate(o, c.Migrate, l).
				Run(context.TODO(), conn.Dialect())
		},
	}
}
