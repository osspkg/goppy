/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"fmt"

	"go.osspkg.com/goppy/v3/orm/dialect"
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
		tag     string
		dialect dialect.Connector
		db      *sql.DB
		err     *atomicError
	}
)

// newStmt init new statement
func newStmt(tag string, dialect dialect.Connector, db *sql.DB, err error) Stmt {
	ae := &atomicError{}
	ae.Set(err)
	return &_stmt{
		tag:     tag,
		dialect: dialect,
		db:      db,
		err:     ae,
	}
}

func (v *_stmt) Close() error {
	v.err.Set(fmt.Errorf("closed connect"))
	return v.db.Close()
}
