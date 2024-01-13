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
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iofile"
	"go.osspkg.com/goppy/sqlcommon"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type (
	ConfigMigrate struct {
		List []ConfigMigrateItem `yaml:"db_migrate"`
	}

	ConfigMigrateItem struct {
		Pool string `yaml:"pool"`
		Dir  string `yaml:"dir"`
	}
)

func (v *ConfigMigrate) Default() {
	if len(v.List) == 0 {
		v.List = []ConfigMigrateItem{
			{
				Pool: "main",
				Dir:  "./migrations",
			},
		}
	}
}

type (
	Migrate interface {
		Run(ctx xc.Context) error
	}
	_migrate struct {
		conn Database
		conf []ConfigMigrateItem
		log  xlog.Logger
	}

	dbm interface {
		Pool(name string) Stmt
		Dialect() string
	}
)

func NewMigrate(conn dbm, conf []ConfigMigrateItem, log xlog.Logger) Migrate {
	return &_migrate{
		conn: conn,
		conf: conf,
		log:  log,
	}
}

func (v *_migrate) Run(ctx xc.Context) error {
	switch v.conn.Dialect() {
	case sqlcommon.MySQLDialect:
		return v.mysql(ctx.Context())
	case sqlcommon.SQLiteDialect:
		return v.sqlite(ctx.Context())
	case sqlcommon.PgSQLDialect:
		return v.pgsql(ctx.Context())
	}
	return nil
}

func hasTable(ctx context.Context, stmt Stmt, query string) bool {
	tables := make([]string, 0)
	err := stmt.QueryContext("check table", ctx, func(q Querier) {
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

func createTable(ctx context.Context, stmt Stmt, query string) error {
	return stmt.ExecContext("create migration table", ctx, func(q Executor) {
		q.SQL(query)
	})
}

func listMigrations(ctx context.Context, stmt Stmt, query string) (map[string]struct{}, error) {
	list := make(map[string]struct{}, 0)
	err := stmt.QueryContext("list migrations", ctx, func(q Querier) {
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

func saveMigration(ctx context.Context, stmt Stmt, query, name string) error {
	return stmt.ExecContext("save migration", ctx, func(q Executor) {
		q.SQL(query, name, time.Now().Unix())
		q.Bind(func(rowsAffected, lastInsertId int64) error {
			if rowsAffected != 1 {
				return fmt.Errorf("cant save migration [%s]", name)
			}
			return nil
		})
	})
}

const (
	pgsqlMigrateList  = `SELECT "name" FROM "__migrations__";`
	pgsqlMigrateSave  = `INSERT INTO "__migrations__" ("name", "timestamp") VALUES ($1, $2);`
	pgsqlMigrateIndex = `CREATE SEQUENCE __migrations___id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1;`
	pgsqlMigrateTable = `CREATE TABLE "__migrations__" (
		"id" integer DEFAULT nextval('__migrations___id_seq') NOT NULL,
		"name" text NOT NULL,
		"timestamp" integer NOT NULL,
		CONSTRAINT "__migrations___pkey" PRIMARY KEY ("id")
	) WITH (oids = false);`
	pgsqlMigrateCheck = `SELECT "tablename" FROM "pg_catalog"."pg_tables" WHERE tablename='__migrations__';`
)

func (v *_migrate) pgsql(ctx context.Context) error {
	return v.executor(ctx, func(stmt Stmt) (map[string]struct{}, error) {
		if !hasTable(ctx, stmt, pgsqlMigrateCheck) {
			if err := createTable(ctx, stmt, pgsqlMigrateIndex); err != nil {
				return nil, err
			}
			if err := createTable(ctx, stmt, pgsqlMigrateTable); err != nil {
				return nil, err
			}
			if !hasTable(ctx, stmt, pgsqlMigrateCheck) {
				return nil, fmt.Errorf("cant create migration table")
			}
		}
		return listMigrations(ctx, stmt, pgsqlMigrateList)
	}, func(stmt Stmt, name string) error {
		return saveMigration(ctx, stmt, pgsqlMigrateSave, name)
	})
}

const (
	mysqlMigrateList  = "SELECT `name` FROM `__migrations__`;"
	mysqlMigrateSave  = "INSERT INTO `__migrations__` (`name`, `timestamp`) VALUES (?, ?);"
	mysqlMigrateTable = "CREATE TABLE `__migrations__` (" +
		"`id` int unsigned NOT NULL AUTO_INCREMENT PRIMARY KEY," +
		"`name` text NOT NULL," +
		"`timestamp` int unsigned NOT NULL" +
		") ENGINE='InnoDB';"
	mysqlMigrateCheck = "SHOW TABLES LIKE '__migrations__';"
)

func (v *_migrate) mysql(ctx context.Context) error {
	return v.executor(ctx, func(stmt Stmt) (map[string]struct{}, error) {
		if !hasTable(ctx, stmt, mysqlMigrateCheck) {
			if err := createTable(ctx, stmt, mysqlMigrateTable); err != nil {
				return nil, err
			}
			if !hasTable(ctx, stmt, mysqlMigrateCheck) {
				return nil, fmt.Errorf("cant create migration table")
			}
		}
		return listMigrations(ctx, stmt, mysqlMigrateList)
	}, func(stmt Stmt, name string) error {
		return saveMigration(ctx, stmt, mysqlMigrateSave, name)
	})
}

const (
	sqliteMigrateList  = "SELECT `name` FROM `__migrations__`;"
	sqliteMigrateSave  = "INSERT INTO `__migrations__` (`name`, `timestamp`) VALUES (?, ?);"
	sqliteMigrateTable = "CREATE TABLE `__migrations__` (" +
		"`id` INTEGER PRIMARY KEY," +
		"`name` text NOT NULL," +
		"`timestamp` TIMESTAMP NOT NULL" +
		");"
	sqliteMigrateCheck = "SELECT `tbl_name` FROM `sqlite_schema` " +
		"WHERE `type` ='table' AND `tbl_name` LIKE '__migrations__';"
)

func (v *_migrate) sqlite(ctx context.Context) error {
	return v.executor(ctx, func(stmt Stmt) (map[string]struct{}, error) {
		if !hasTable(ctx, stmt, sqliteMigrateCheck) {
			if err := createTable(ctx, stmt, sqliteMigrateTable); err != nil {
				return nil, err
			}
			if !hasTable(ctx, stmt, sqliteMigrateCheck) {
				return nil, fmt.Errorf("cant create migration table")
			}
		}
		return listMigrations(ctx, stmt, sqliteMigrateList)
	}, func(stmt Stmt, name string) error {
		return saveMigration(ctx, stmt, sqliteMigrateSave, name)
	})
}

func (v *_migrate) executor(ctx context.Context,
	call func(stmt Stmt) (map[string]struct{}, error),
	save func(stmt Stmt, name string) error,
) error {
	for _, migrateItem := range v.conf {
		if !iofile.Exist(migrateItem.Dir) {
			continue
		}

		stmt := v.conn.Pool(migrateItem.Pool)

		exist, err := call(stmt)
		if err != nil {
			return err
		}

		list, err := filepath.Glob(migrateItem.Dir + "/*.sql")
		if err != nil {
			return errors.Wrapf(err, "get migration files")
		}
		sort.Strings(list)
		for _, f := range list {
			name := filepath.Base(f)
			if _, ok := exist[name]; ok {
				continue
			}
			v.log.WithFields(xlog.Fields{"file": f}).Infof("new migration")

			b, err0 := os.ReadFile(f)
			if err0 != nil {
				return errors.Wrapf(err0, "read migration file [%s]", name)
			}
			if err = stmt.ExecContext("new migration", ctx, func(q Executor) {
				q.SQL(string(b))
			}); err != nil {
				return errors.Wrapf(err, "exec migration file [%s]", name)
			}

			if err = save(stmt, name); err != nil {
				return errors.Wrapf(err, "save migrated file [%s]", name)
			}
		}
	}
	return nil
}
