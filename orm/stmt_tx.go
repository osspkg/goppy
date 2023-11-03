/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
)

var poolTx = sync.Pool{New: func() interface{} { return &tx{} }}

type (
	Tx interface {
		Exec(vv ...func(e Executor))
		Query(vv ...func(q Querier))
	}

	tx struct {
		v []interface{}
	}

	dbGetter interface {
		QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
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

func (v *tx) Reset() *tx {
	v.v = v.v[:0]
	return v
}

func (s *_stmt) TransactionContext(name string, ctx context.Context, call func(v Tx)) error {
	q, ok := poolTx.Get().(*tx)
	if !ok {
		return errInvalidModelPool
	}
	defer poolTx.Put(q.Reset())

	call(q)

	return s.TxContext(name, ctx, func(ctx context.Context, tx *sql.Tx) error {
		for i, c := range q.v {
			if cc, ok := c.(func(q Executor)); ok {
				if err := callExecContext(ctx, tx, cc, s.db.Dialect()); err != nil {
					return err
				}
				continue
			}
			if cc, ok := c.(func(q Querier)); ok {
				if err := callQueryContext(ctx, tx, cc); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("unknown query model #%d", i)
		}
		return nil
	})
}
