/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"

	"go.osspkg.com/goppy/sdk/errors"
)

var (
	errInvalidModelPool = errors.New("invalid decoder pool")
)

type (
	//Stmt statement model
	Stmt interface {
		Ping() error
		CallContext(name string, ctx context.Context, callFunc func(context.Context, *sql.DB) error) error
		TxContext(name string, ctx context.Context, callFunc func(context.Context, *sql.Tx) error) error

		ExecContext(name string, ctx context.Context, call func(q Executor)) error
		QueryContext(name string, ctx context.Context, call func(q Querier)) error
		TransactionContext(name string, ctx context.Context, call func(v Tx)) error
	}

	_stmt struct {
		name string
		db   dbPool
		opts *options
	}

	dbPool interface {
		Dialect() string
		Pool(string) (*sql.DB, error)
	}
)

// newStmt init new statement
func newStmt(name string, db dbPool, p *options) Stmt {
	return &_stmt{
		name: name,
		db:   db,
		opts: p,
	}
}
