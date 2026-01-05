/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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

	"go.osspkg.com/goppy/v3/orm/dialect"
)

type Migrator interface {
	Run(ctx context.Context) error
}

type migrate struct {
	conn ORM
	fs   FS
	mux  sync.Mutex
}

func NewMigrate(o ORM, fs FS) Migrator {
	return &migrate{
		conn: o,
		fs:   fs,
	}
}

func (v *migrate) Run(ctx context.Context) error {
	return v.executor(ctx,
		func(dialectName dialect.Name, stmt Stmt, mig dialect.Migrator) (map[string]struct{}, error) {
			if !migrateTableCheck(ctx, stmt, mig.CheckTableQuery()) {
				for _, createQuery := range mig.CreateTableQuery() {
					if err := migrateCreateTable(ctx, stmt, createQuery); err != nil {
						return nil, fmt.Errorf("create table for '%s': %w", dialectName, err)
					}
				}
			}

			return migrateCompletedList(ctx, stmt, mig.CompletedQuery())
		},
		func(name string, stmt Stmt, mig dialect.Migrator) error {
			return migrateSave(ctx, stmt, mig.SaveQuery(), name)
		},
	)
}

func (v *migrate) executor(
	ctx context.Context,
	getExists func(dialectName dialect.Name, stmt Stmt, mig dialect.Migrator) (map[string]struct{}, error),
	saveNew func(name string, stmt Stmt, mig dialect.Migrator) error,
) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	defer v.fs.Done()
	for v.fs.Next() {

		dialectName := v.fs.Dialect()

		mig, ok := dialect.GetMigrator(dialectName)
		if !ok {
			continue
		}

		for _, tag := range v.fs.Tags() {

			stmt := v.conn.Tag(tag)
			exist, err := getExists(dialectName, stmt, mig)
			if err != nil {
				return fmt.Errorf("get completed migration for tag '%s:%s': %w", dialectName, tag, err)
			}

			list, err := v.fs.FileNames()
			if err != nil {
				return fmt.Errorf("get migration files for tag '%s:%s': %w", dialectName, tag, err)
			}

			for _, filePath := range list {
				name := filepath.Base(filePath)
				if _, ok := exist[name]; ok {
					continue
				}

				rawData, err := v.fs.FileData(filePath)
				if err != nil {
					logx.Error("New DB migration", "dialect", dialectName, "tag", tag,
						"file", filePath, "err", err)

					return fmt.Errorf("read migration file '%s' for tag '%s:%s': %w ", name, dialectName, tag, err)
				}

				for _, subsql := range strings.Split(rawData, ";") {
					subsql = removeSQLComment(subsql)
					if len(subsql) == 0 {
						continue
					}

					if err = stmt.Exec(ctx, "new migration", func(q Executor) { q.SQL(subsql) }); err != nil {
						logx.Error("New DB migration", "dialect", dialectName, "tag", tag,
							"file", filePath, "sql", subsql, "err", err)

						return errors.Wrapf(err, "exec migration file '%s', sql: '%s'", name, subsql)
					}
				}

				if err = saveNew(name, stmt, mig); err != nil {
					logx.Error("New DB migration", "dialect", dialectName, "tag", tag,
						"file", filePath, "err", err)

					return errors.Wrapf(err, "save migrated file '%s'", name)
				}

				logx.Info("New DB migration", "dialect", dialectName, "tag", tag, "file", filePath)
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
