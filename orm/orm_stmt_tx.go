/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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
		Exec(vv ...func(e Executor))
		Query(vv ...func(q Querier))
	}

	tx struct {
		v []any
	}

	dbGetter interface {
		QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	}
)

func (v *tx) Exec(vv ...func(q Executor)) {
	for _, f := range vv {
		v.v = append(v.v, f)
	}
}

func (v *tx) Query(vv ...func(q Querier)) {
	for _, f := range vv {
		v.v = append(v.v, f)
	}
}

func (v *tx) Reset() {
	v.v = v.v[:0]
}

func (v *_stmt) Tx(ctx context.Context, name string, call func(v Tx)) error {
	q := poolTx.Get()
	defer poolTx.Put(q)

	call(q)

	return v.TxContext(ctx, name, func(ctx context.Context, db DB) error {
		for i, c := range q.v {
			if cc, ok := c.(func(q Executor)); ok {
				if err := callExecContext(ctx, db, cc, v.dialect); err != nil {
					return err
				}
				continue
			}
			if cc, ok := c.(func(q Querier)); ok {
				if err := callQueryContext(ctx, db, cc); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("unknown query model #%d", i)
		}
		return nil
	})
}
