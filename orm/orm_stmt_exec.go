/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"

	"go.osspkg.com/ioutils/pool"
)

var poolExec = pool.New[*exec](func() *exec { return &exec{} })

type exec struct {
	D string
	Q string
	P [][]any
	B func(rowsAffected, lastInsertId int64) error
}

func (v *exec) SQL(query string, args ...any) {
	v.Q = query
	v.Params(args...)
}

func (v *exec) Params(args ...any) {
	if len(args) == 0 {
		return
	}

	switch v.D {
	case PgSQLDialect:
		applyPGSqlCastTypes(args)
	default:
	}

	v.P = append(v.P, args)
}

func (v *exec) Bind(call func(rowsAffected, lastInsertId int64) error) {
	v.B = call
}

func (v *exec) Reset() {
	v.Q, v.P, v.B, v.D = "", v.P[:0], nil, ""
}

type (
	// Executor interface for generate execute query
	Executor interface {
		SQL(query string, args ...any)
		Params(args ...any)
		Bind(call func(rowsAffected, lastInsertId int64) error)
	}
)

func (v *_stmt) Exec(ctx context.Context, name string, call func(q Executor)) error {
	return v.CallContext(ctx, name, func(ctx context.Context, db DB) error {
		return callExecContext(ctx, db, call, v.dialect)
	})
}

func callExecContext(ctx context.Context, db dbGetter, call func(q Executor), dialect string) error {
	q := poolExec.Get()
	defer func() { poolExec.Put(q) }()

	q.D = dialect

	call(q)

	if len(q.P) == 0 {
		q.P = append(q.P, []any{})
	}
	stmt, err := db.PrepareContext(ctx, q.Q)
	if err != nil {
		return err
	}
	defer stmt.Close() // nolint: errcheck
	var rowsAffected, lastInsertId int64
	for _, params := range q.P {
		result, err0 := stmt.ExecContext(ctx, params...)
		if err0 != nil {
			return err0
		}
		rows, err0 := result.RowsAffected()
		if err0 != nil {
			return err0
		}
		rowsAffected += rows

		if dialect != PgSQLDialect {
			rows, err0 = result.LastInsertId()
			if err0 != nil {
				return err0
			}
			lastInsertId = rows
		}
	}
	if err = stmt.Close(); err != nil {
		return err
	}
	if q.B == nil {
		return nil
	}
	return q.B(rowsAffected, lastInsertId)
}
