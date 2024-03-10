/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ormsqlite

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iofile"
	"go.osspkg.com/goppy/orm"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/routine"
	"go.osspkg.com/goppy/sqlcommon"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

// ConfigSqlite sqlite config model
type ConfigSqlite struct {
	Pool    []Item                  `yaml:"sqlite"`
	Migrate []orm.ConfigMigrateItem `yaml:"sqlite_migrate"`
}

func (v *ConfigSqlite) Default() {
	if len(v.Pool) == 0 {
		v.Pool = []Item{
			{
				Name:        "main",
				File:        "./sqlite.db",
				Cache:       "private",
				Mode:        "rwc",
				Journal:     "WAL",
				LockingMode: "EXCLUSIVE",
				OtherParams: "auto_vacuum=incremental",
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

// List getting all configs
func (v *ConfigSqlite) List() (list []sqlcommon.ItemInterface) {
	for _, vv := range v.Pool {
		list = append(list, vv)
	}
	return
}

// WithClient launch SQLite connection pool
func WithClient() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigSqlite{},
		Inject: func(c *ConfigSqlite, l xlog.Logger) SQLite {
			conn := New(c)
			o := orm.New(conn, orm.UsePluginLogger(l))
			m := orm.NewMigrate(o, c.Migrate, l)
			return &sqliteProvider{
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
	sqliteProvider struct {
		conn    sqlcommon.Connector
		orm     orm.Database
		migrate orm.Migrate
		conf    ConfigSqlite
		log     xlog.Logger
		list    map[string]orm.Stmt
		mux     sync.RWMutex
		active  bool
	}

	// SQLite connection SQLite interface
	SQLite interface {
		Pool(name string) orm.Stmt
	}
)

func (v *sqliteProvider) Up(ctx xc.Context) error {
	routine.Interval(ctx.Context(), time.Second*5, func(_ context.Context) {
		var recovery bool
		if v.active {
			v.mux.RLock()
			for name, stmt := range v.list {
				if err := stmt.Ping(); err != nil {
					v.log.WithFields(
						xlog.Fields{"err": fmt.Errorf("pool `%s`: %w", name, err).Error()},
					).Errorf("SQLite check connect")
					v.active = false
				}
			}
			v.mux.RUnlock()

			for _, item := range v.conf.Pool {
				if !iofile.Exist(item.File) {
					v.log.WithFields(
						xlog.Fields{"err": fmt.Sprintf("pool `%s`: [%s] file is missing", item.Name, item.File)},
					).Errorf("SQLite check connect")
					v.active = false
					recovery = true
				}
			}
		}

		if !v.active {
			if err := v.updateConnect(); err == nil {
				v.updateList()
				v.active = true

				if recovery {
					v.log.Infof("SQLite recovery migration")
					if err = v.migrate.Run(ctx); err != nil {
						v.log.WithFields(
							xlog.Fields{"err": err.Error()},
						).Errorf("SQLite recovery migration")
					}
				}
			} else {
				v.log.WithFields(
					xlog.Fields{"err": err.Error()},
				).Errorf("SQLite update connections")
			}
		}
	})
	if !v.active {
		return errors.New("Failed to connect to database")
	}
	return v.migrate.Run(ctx)
}

func (v *sqliteProvider) Down() error {
	return v.conn.Close()
}

func (v *sqliteProvider) updateList() {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, vv := range v.conf.Pool {
		v.list[vv.Name] = v.orm.Pool(vv.Name)
	}
}

func (v *sqliteProvider) updateConnect() error {
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
			xlog.Fields{vv.Name: vv.File},
		).Infof("SQLite update connections")
	}
	return errs
}

func (v *sqliteProvider) Pool(name string) orm.Stmt {
	v.mux.RLock()
	defer v.mux.RUnlock()
	if s, ok := v.list[name]; ok {
		return s
	}
	return v.orm.Pool(name)
}

func (v *sqliteProvider) Dialect() string {
	return v.orm.Dialect()
}
