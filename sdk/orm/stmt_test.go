/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/osspkg/goppy/sdk/orm"
	"github.com/osspkg/goppy/sdk/orm/plugins"
	"github.com/osspkg/goppy/sdk/orm/schema/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Stmt(t *testing.T) {
	file, err := os.CreateTemp("/tmp", "prefix")
	require.NoError(t, err)
	defer os.Remove(file.Name()) //nolint: errcheck

	conn := sqlite.New(&sqlite.Config{Pool: []sqlite.Item{{Name: "main", File: file.Name()}}})
	require.NoError(t, conn.Reconnect())
	defer conn.Close() //nolint: errcheck
	pool := orm.New(conn,
		orm.UsePluginLogger(plugins.StdOutLog),
		orm.UsePluginMetric(plugins.StdOutMetric),
	).Pool("main")

	err = pool.CallContext("init", context.Background(), func(ctx context.Context, db *sql.DB) error {
		sqls := []string{
			`create table users (
				id		INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
				name	TEXT
			);`,
			"insert into `users` (`id`, `name`) values (1, 'aaaa');",
			"insert into `users` (`id`, `name`) values (2, 'bbbb');",
		}

		for _, item := range sqls {
			if _, err = db.ExecContext(ctx, item); err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(t, err)

	err = pool.QueryContext("", context.Background(), func(q orm.Querier) {
		q.SQL("select `name` from `users` where `id` = ?", 1)
		q.Bind(func(bind orm.Scanner) error {
			name := ""
			assert.NoError(t, bind.Scan(&name))
			assert.Equal(t, "aaaa", name)
			return nil
		})
	})
	assert.NoError(t, err)

	var result []string
	err = pool.QueryContext("", context.Background(), func(q orm.Querier) {
		q.SQL("select `name` from `users`")
		q.Bind(func(bind orm.Scanner) error {
			name := ""
			assert.NoError(t, bind.Scan(&name))
			result = append(result, name)
			return nil
		})
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"aaaa", "bbbb"}, result)

	err = pool.ExecContext("", context.Background(), func(e orm.Executor) {
		e.SQL("insert into `users` (`id`, `name`) values (?, ?);")
		e.Params(3, "cccc")
		e.Params(4, "dddd")

		e.Bind(func(rowsAffected, lastInsertId int64) error {
			assert.Equal(t, int64(2), rowsAffected)
			assert.Equal(t, int64(4), lastInsertId)
			return nil
		})
	})
	assert.NoError(t, err)

	var result2 []string
	err = pool.QueryContext("", context.Background(), func(q orm.Querier) {
		q.SQL("select `name` from `users`")
		q.Bind(func(bind orm.Scanner) error {
			name := ""
			err = bind.Scan(&name)
			result2 = append(result2, name)
			return err
		})
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"aaaa", "bbbb", "cccc", "dddd"}, result2)

	var result3 []string
	err = pool.TransactionContext("", context.Background(), func(v orm.Tx) {
		v.Exec(func(e orm.Executor) {
			e.SQL("insert into `users` (`id`, `name`) values (?, ?);")
			e.Params(10, "abcd")
			e.Params(11, "efgh")
			e.Bind(func(rowsAffected, lastInsertId int64) error {
				assert.Equal(t, int64(2), rowsAffected)
				assert.Equal(t, int64(11), lastInsertId)
				return nil
			})
		})
		v.Query(func(q orm.Querier) {
			q.SQL("select `name` from `users`")
			q.Bind(func(bind orm.Scanner) error {
				name := ""
				err = bind.Scan(&name)
				result3 = append(result3, name)
				return err
			})
		})
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"aaaa", "bbbb", "cccc", "dddd", "abcd", "efgh"}, result3)

	var result4 []string
	err = pool.QueryContext("", context.Background(), func(q orm.Querier) {
		q.SQL("select `name` from `users`")
		q.Bind(func(bind orm.Scanner) error {
			name := ""
			err = bind.Scan(&name)
			result4 = append(result4, name)
			return err
		})
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{"aaaa", "bbbb", "cccc", "dddd", "abcd", "efgh"}, result4)
}

func BenchmarkStmt(b *testing.B) {
	file, err := os.CreateTemp("/tmp", "prefix")
	require.NoError(b, err)
	defer os.Remove(file.Name()) //nolint: errcheck

	conn := sqlite.New(&sqlite.Config{Pool: []sqlite.Item{{Name: "main", File: file.Name()}}})
	require.NoError(b, conn.Reconnect())
	defer conn.Close() //nolint: errcheck
	pool := orm.New(conn).Pool("main")

	err = pool.CallContext("init", context.Background(), func(ctx context.Context, db *sql.DB) error {
		sqls := []string{
			`create table users (
				id		INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
				name	TEXT
			);`,
		}

		for _, item := range sqls {
			if _, err = db.ExecContext(ctx, item); err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(b, err)

	b.Run("insert", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 1; i < b.N; i++ {
			err = pool.ExecContext("", context.Background(), func(e orm.Executor) {
				i := i
				e.SQL("insert or ignore into `users` (`id`, `name`) values (?, ?);")
				e.Params(i, "cccc")
			})
			assert.NoError(b, err)
		}
	})

	var name string
	b.Run("select", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 1; i < b.N; i++ {
			err = pool.QueryContext("", context.Background(), func(q orm.Querier) {
				i := i
				q.SQL("select `name` from `users` where `id` = ?", i)
				q.Bind(func(bind orm.Scanner) error {
					return bind.Scan(&name)
				})
			})
			assert.NoError(b, err)
		}
	})
}
