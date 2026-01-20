/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"fmt"

	"go.osspkg.com/ioutils/pool"
)

var poolTx = pool.New[*tx](func() *tx { return &tx{} })

type (
	Tx interface {
		Exec(args ...func(e Executor))
		Query(args ...func(q Querier))
	}

	tx struct {
		funcs []any
	}

	queryGetter interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	}
)

func (v *tx) Exec(args ...func(q Executor)) {
	if v.funcs == nil {
		v.funcs = make([]any, 0, len(args))
	}

	for _, f := range args {
		v.funcs = append(v.funcs, f)
	}
}

func (v *tx) Query(args ...func(q Querier)) {
	if v.funcs == nil {
		v.funcs = make([]any, 0, len(args))
	}

	for _, f := range args {
		v.funcs = append(v.funcs, f)
	}
}

func (v *tx) Reset() {
	v.funcs = v.funcs[:0]
}

func (v *_stmt) Tx(ctx context.Context, name string, call func(v Tx)) error {
	obj := poolTx.Get()
	defer poolTx.Put(obj)

	call(obj)

	return v.TxContext(ctx, name, func(ctx context.Context, db DB) error {
		for i, cb := range obj.funcs {

			if ex, ok := cb.(func(q Executor)); ok {
				if err := callExecContext(ctx, db, ex, v.dialect); err != nil {
					return err
				}
				continue
			}

			if qu, ok := cb.(func(q Querier)); ok {
				if err := callQueryContext(ctx, db, qu, v.dialect); err != nil {
					return err
				}
				continue
			}

			return fmt.Errorf("unknown query func #%d", i)
		}

		return nil
	})
}
