/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormmysql

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/orm"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/routine"
	"go.osspkg.com/goppy/sqlcommon"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

// ConfigMysql mysql config model
type ConfigMysql struct {
	Pool    []Item                  `yaml:"mysql"`
	Migrate []orm.ConfigMigrateItem `yaml:"mysql_migrate"`
}

// List getting all configs
func (v *ConfigMysql) List() (list []sqlcommon.ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

func (v *ConfigMysql) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []Item{
			{
				Name:              "main",
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
		v.Migrate = []orm.ConfigMigrateItem{
			{
				Pool: "main",
				Dir:  "./migrations",
			},
		}
	}
}

// WithClient launch MySQL connection pool
func WithClient() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigMysql{},
		Inject: func(c *ConfigMysql, l xlog.Logger) MySQL {
			conn := New(c)
			o := orm.New(conn, orm.UsePluginLogger(l))
			m := orm.NewMigrate(o, c.Migrate, l)
			return &mysqlProvider{
				conn:    conn,
				orm:     o,
				migrate: m,
				conf:    *c,
				log:     l,
				list:    make(map[string]orm.Stmt),
				active:  false,
			}
		},
	}
}

type (
	mysqlProvider struct {
		conn    sqlcommon.Connector
		orm     orm.Database
		migrate orm.Migrate
		conf    ConfigMysql
		log     xlog.Logger
		list    map[string]orm.Stmt
		mux     sync.RWMutex
		active  bool
	}

	// MySQL connection MySQL interface
	MySQL interface {
		Pool(name string) orm.Stmt
	}
)

func (v *mysqlProvider) Up(ctx xc.Context) error {
	routine.Interval(ctx.Context(), time.Second*5, func(ctx context.Context) {
		if v.active {
			v.mux.RLock()
			for name, stmt := range v.list {
				if err := stmt.Ping(); err != nil {
					v.log.WithFields(
						xlog.Fields{"err": fmt.Errorf("pool `%s`: %w", name, err).Error()},
					).Errorf("MySQL check connect")
					v.active = false
				}
			}
			v.mux.RUnlock()
		}

		if !v.active {
			if err := v.updateConnect(); err == nil {
				v.updateList()
				v.active = true
			} else {
				v.log.WithFields(
					xlog.Fields{"err": err.Error()},
				).Errorf("MySQL update connections")
			}
		}
	})
	if !v.active {
		return errors.New("Failed to connect to database")
	}
	return v.migrate.Run(ctx)
}

func (v *mysqlProvider) Down() error {
	return v.conn.Close()
}

func (v *mysqlProvider) updateList() {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, vv := range v.conf.Pool {
		v.list[vv.Name] = v.orm.Pool(vv.Name)
	}
}

func (v *mysqlProvider) updateConnect() error {
	if err := v.conn.Reconnect(); err != nil {
		return err
	}
	var errs error
	for _, vv := range v.conf.Pool {
		p, err := v.conn.Pool(vv.Name)
		if err != nil {
			errs = errors.Wrap(errs, fmt.Errorf("pool `%s`: %w", vv.Name, err))
			continue
		}
		if err = p.Ping(); err != nil {
			errs = errors.Wrap(errs, fmt.Errorf("pool `%s`: %w", vv.Name, err))
			continue
		}
		v.log.WithFields(
			xlog.Fields{vv.Name: fmt.Sprintf("%s:%d", vv.Host, vv.Port)},
		).Infof("MySQL update connections")
	}
	return errs
}

func (v *mysqlProvider) Pool(name string) orm.Stmt {
	v.mux.RLock()
	defer v.mux.RUnlock()
	if s, ok := v.list[name]; ok {
		return s
	}
	return v.orm.Pool(name)
}

func (v *mysqlProvider) Dialect() string {
	return v.orm.Dialect()
}
