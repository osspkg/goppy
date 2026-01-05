/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"

	"go.osspkg.com/ioutils/pool"

	"go.osspkg.com/goppy/v3/orm/dialect"
)

var poolExec = pool.New[*exec](func() *exec { return &exec{} })

type exec struct {
	Dialect dialect.Connector
	Query   string
	Args    [][]any
	Binding func(rowsAffected, lastInsertId int64) error
}

func (v *exec) SQL(query string, args ...any) {
	v.Query = query
	v.Params(args...)
}

func (v *exec) Params(args ...any) {
	if len(args) == 0 {
		return
	}

	if cast := v.Dialect.CastTypesFunc(); cast != nil {
		cast(args)
	}

	v.Args = append(v.Args, args)
}

func (v *exec) Bind(call func(rowsAffected, lastInsertId int64) error) {
	v.Binding = call
}

func (v *exec) Reset() {
	v.Query, v.Args, v.Binding, v.Dialect = "", v.Args[:0], nil, nil
}

type (
	// Executor interface for generate execute query
	Executor interface {
		SQL(query string, args ...any)
		Params(args ...any)
		Bind(call func(rowsAffected, lastInsertId int64) error)
	}
)

func (v *_stmt) Exec(ctx context.Context, name string, call func(e Executor)) error {
	return v.CallContext(ctx, name, func(ctx context.Context, db DB) error {
		return callExecContext(ctx, db, call, v.dialect)
	})
}

func callExecContext(ctx context.Context, db queryGetter, call func(e Executor), dc dialect.Connector) error {
	obj := poolExec.Get()
	defer func() { poolExec.Put(obj) }()

	obj.Dialect = dc

	call(obj)

	if len(obj.Args) == 0 {
		obj.Args = append(obj.Args, []any{})
	}

	stmt, err := db.PrepareContext(ctx, obj.Query)
	if err != nil {
		return err
	}
	defer stmt.Close() // nolint: errcheck

	var rowsAffected, lastInsertId int64
	for _, params := range obj.Args {
		result, err0 := stmt.ExecContext(ctx, params...)
		if err0 != nil {
			return fmt.Errorf("failed exec: %w", err0)
		}

		rows, err0 := result.RowsAffected()
		if err0 != nil {
			return fmt.Errorf("failed get row affected: %w", err0)
		}

		rowsAffected += rows

		if obj.Dialect.HasLastInsertId() {
			rows, err0 = result.LastInsertId()
			if err0 != nil {
				return fmt.Errorf("failed get last insert id: %w", err0)
			}

			lastInsertId = rows
		}
	}

	if err = stmt.Close(); err != nil {
		return fmt.Errorf("failed close: %w", err)
	}

	if obj.Binding == nil {
		return nil
	}

	if err = obj.Binding(rowsAffected, lastInsertId); err != nil {
		return fmt.Errorf("failed bind: %w", err)
	}

	return nil
}
