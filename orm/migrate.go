/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
)

type (
	ConfigMigrate struct {
		List []ConfigMigrateItem `yaml:"db_migrate"`
	}

	ConfigMigrateItem struct {
		Tags    string `yaml:"tags"`
		Dialect string `yaml:"dialect"`
		Dir     string `yaml:"dir"`
	}
)

func (v *ConfigMigrate) Default() {
	if len(v.List) == 0 {
		v.List = []ConfigMigrateItem{
			{
				Tags:    "master",
				Dialect: MySQLDialect,
				Dir:     "./migrations/" + MySQLDialect,
			},
			{
				Tags:    "master",
				Dialect: PgSQLDialect,
				Dir:     "./migrations/" + PgSQLDialect,
			},
			{
				Tags:    "master",
				Dialect: SQLiteDialect,
				Dir:     "./migrations/" + SQLiteDialect,
			},
		}
	}
}

// ---------------------------------------------------------------------------------------------------------------------

type Migrate struct {
	Conn ORM
	FS   MFS
	mux  sync.Mutex
}

func (v *Migrate) Run(ctx context.Context, dialect string) error {
	return v.executor(ctx, dialect,
		func(stmt Stmt, mig Migrator) (map[string]struct{}, error) {
			if !migrateTableCheck(ctx, stmt, mig.CheckTableQuery()) {
				for _, createQuery := range mig.CreateTableQuery() {
					if err := migrateCreateTable(ctx, stmt, createQuery); err != nil {
						return nil, err
					}
				}
				if !migrateTableCheck(ctx, stmt, mig.CheckTableQuery()) {
					return nil, fmt.Errorf("cant create migration table")
				}
			}
			return migrateCompletedList(ctx, stmt, mig.CompletedQuery())
		},
		func(name string, stmt Stmt, mig Migrator) error {
			return migrateSave(ctx, stmt, mig.SaveQuery(), name)
		},
	)
}

func (v *Migrate) executor(ctx context.Context, dialect string,
	call func(stmt Stmt, mig Migrator) (map[string]struct{}, error),
	save func(name string, stmt Stmt, mig Migrator) error,
) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	defer v.FS.Done()

	for v.FS.Next() {
		if dialect != v.FS.Dialect() {
			continue
		}

		mig, err := dialectExtract(v.FS.Dialect())
		if err != nil {
			return err
		}

		for _, tag := range v.FS.Tags() {
			stmt := v.Conn.Tag(tag)
			exist, err := call(stmt, mig)
			if err != nil {
				return err
			}
			list, err := v.FS.FileNames()
			if err != nil {
				return errors.Wrapf(err, "get migration files")
			}
			for _, filePath := range list {
				name := filepath.Base(filePath)
				if _, ok := exist[name]; ok {
					continue
				}
				logx.Info("New DB migration", "dialect", dialect, "tag", tag, "file", filePath)
				sqldata, err := v.FS.FileData(filePath)
				if err != nil {
					return errors.Wrapf(err, "read migration file [%s]", name)
				}
				for _, subsql := range strings.Split(sqldata, ";") {
					subsql = removeSQLComment(subsql)
					if len(subsql) == 0 {
						continue
					}
					if err = stmt.Exec(ctx, "new migration", func(q Executor) {
						q.SQL(subsql)
					}); err != nil {
						return errors.Wrapf(err, "exec migration file [%s], sql: `%s`", name, subsql)
					}
				}
				if err = save(name, stmt, mig); err != nil {
					return errors.Wrapf(err, "save migrated file [%s]", name)
				}
			}
		}

	}
	return nil
}

var sqlCommentRex = regexp.MustCompile(`(?mU)--.*\n`)

func removeSQLComment(s string) string {
	s = sqlCommentRex.ReplaceAllString(s, "\n")
	return strings.TrimSpace(s)
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
