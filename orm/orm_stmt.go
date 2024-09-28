/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"fmt"
)

type (
	// Stmt statement model
	Stmt interface {
		PingContext(ctx context.Context) error
		CallContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) error
		TxContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) error

		Exec(ctx context.Context, name string, call func(q Executor)) error
		Query(ctx context.Context, name string, call func(q Querier)) error
		Tx(ctx context.Context, name string, call func(v Tx)) error

		Close() error
	}

	DB interface {
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
		ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	}

	_stmt struct {
		dialect string
		conn    *sql.DB
		opts    *options
		err     error
	}
)

// newStmt init new statement
func newStmt(d string, c *sql.DB, p *options, e error) Stmt {
	return &_stmt{dialect: d, conn: c, opts: p, err: e}
}

func (v *_stmt) Close() error {
	v.err = fmt.Errorf("closed connect")
	return v.conn.Close()
}