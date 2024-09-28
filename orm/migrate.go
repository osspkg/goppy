/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
)

var (
	migrateQuery = make(map[string]Migrator, 10)
	mLock        = syncing.NewLock()
)

func Register(dialog string, migrate Migrator) {
	mLock.Lock(func() {
		migrateQuery[dialog] = migrate
	})
}

func MigrateProvider(dialog string) (migrate Migrator, err error) {
	mLock.Lock(func() {
		m, ok := migrateQuery[dialog]
		if ok {
			migrate = m
			return
		}
		err = fmt.Errorf("migrate query not found")
	})
	return
}

// ---------------------------------------------------------------------------------------------------------------------

type (
	ConfigMigrate struct {
		List []ConfigMigrateItem `yaml:"db_migrate"`
	}

	ConfigMigrateItem struct {
		Tags string `yaml:"tags"`
		Dir  string `yaml:"dir"`
	}
)

func (v *ConfigMigrate) Default() {
	if len(v.List) == 0 {
		v.List = []ConfigMigrateItem{
			{
				Tags: "master",
				Dir:  "./migrations",
			},
		}
	}
}

// ---------------------------------------------------------------------------------------------------------------------

type (
	Migrate struct {
		conn ORM
		conf []ConfigMigrateItem
	}
)

func NewMigrate(conn ORM, conf []ConfigMigrateItem) *Migrate {
	return &Migrate{
		conn: conn,
		conf: conf,
	}
}

func (v *Migrate) Run(ctx context.Context, dialect string) error {
	mig, err := MigrateProvider(dialect)
	if err != nil {
		return err
	}
	return v.executor(ctx, func(stmt Stmt) (map[string]struct{}, error) {
		if !migrateTableCheck(ctx, stmt, mig.CheckTableQuery()) {
			for _, createQuery := range mig.CreateTableQuery() {
				if err = migrateCreateTable(ctx, stmt, createQuery); err != nil {
					return nil, err
				}
			}
			if !migrateTableCheck(ctx, stmt, mig.CheckTableQuery()) {
				return nil, fmt.Errorf("cant create migration table")
			}
		}
		return migrateCompletedList(ctx, stmt, mig.CompletedQuery())
	}, func(stmt Stmt, name string) error {
		return migrateSave(ctx, stmt, mig.SaveQuery(), name)
	})
}

func migrateTableCheck(ctx context.Context, stmt Stmt, query string) bool {
	tables := make([]string, 0)
	err := stmt.Query(ctx, "check table", func(q Querier) {
		q.SQL(query)
		q.Bind(func(bind Scanner) error {
			var table string
			if err := bind.Scan(&table); err != nil {
				return err
			}
			tables = append(tables, table)
			return nil
		})
	})
	if err != nil {
		return false
	}
	return len(tables) == 1
}

func migrateCreateTable(ctx context.Context, stmt Stmt, query string) error {
	return stmt.Exec(ctx, "create migration table", func(q Executor) {
		q.SQL(query)
	})
}

func migrateCompletedList(ctx context.Context, stmt Stmt, query string) (map[string]struct{}, error) {
	list := make(map[string]struct{}, 0)
	err := stmt.Query(ctx, "list migrations", func(q Querier) {
		q.SQL(query)
		q.Bind(func(bind Scanner) error {
			var name string
			if err := bind.Scan(&name); err != nil {
				return err
			}
			list[name] = struct{}{}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func migrateSave(ctx context.Context, stmt Stmt, query, name string) error {
	return stmt.Exec(ctx, "save migration", func(q Executor) {
		q.SQL(query, name, time.Now().Unix())
		q.Bind(func(rowsAffected, _ int64) error {
			if rowsAffected != 1 {
				return fmt.Errorf("cant save migration [%s]", name)
			}
			return nil
		})
	})
}

func (v *Migrate) executor(ctx context.Context,
	call func(stmt Stmt) (map[string]struct{}, error),
	save func(stmt Stmt, name string) error,
) error {
	for _, migrateItem := range v.conf {
		if !fs.FileExist(migrateItem.Dir) {
			continue
		}
		for _, tag := range strings.Split(migrateItem.Tags, ",") {
			stmt := v.conn.Tag(tag)
			exist, err := call(stmt)
			if err != nil {
				return err
			}
			list, err := filepath.Glob(migrateItem.Dir + "/*.sql")
			if err != nil {
				return errors.Wrapf(err, "get migration files")
			}
			sort.Strings(list)
			for _, filePath := range list {
				name := filepath.Base(filePath)
				if _, ok := exist[name]; ok {
					continue
				}
				logx.Info("New migration", "file", filePath)
				b, err0 := os.ReadFile(filePath)
				if err0 != nil {
					return errors.Wrapf(err0, "read migration file [%s]", name)
				}
				if err = stmt.Exec(ctx, "new migration", func(q Executor) {
					q.SQL(string(b))
				}); err != nil {
					return errors.Wrapf(err, "exec migration file [%s]", name)
				}
				if err = save(stmt, name); err != nil {
					return errors.Wrapf(err, "save migrated file [%s]", name)
				}
			}
		}

	}
	return nil
}
