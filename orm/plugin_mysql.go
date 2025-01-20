/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"time"

	"go.osspkg.com/goppy/v2/plugins"
)

// ConfigMysql mysql config model
type ConfigMysql struct {
	Pool []ConfigMysqlClient `yaml:"mysql"`
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
}

// WithMysqlClient launch MySQL connection pool
func WithMysqlClient() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigMysql{},
		Inject: func(c *ConfigMysql, o ORM) error {
			conn := NewMysqlClient(c)
			return o.Register(conn, func() {
				dialectRegister(MySQLDialect, &_mysqlMigrateTable{})
			})
		},
	}
}
