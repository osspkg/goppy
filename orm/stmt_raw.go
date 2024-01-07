/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"

	"go.osspkg.com/goppy/errors"
)

// Ping database ping
func (s *_stmt) Ping() error {
	return s.CallContext("ping", context.Background(), func(ctx context.Context, db *sql.DB) error {
		return db.PingContext(ctx)
	})
}

// CallContext basic query execution
func (s *_stmt) CallContext(name string, ctx context.Context, callFunc func(context.Context, *sql.DB) error) error {
	pool, err := s.db.Pool(s.name)
	if err != nil {
		return err
	}

	s.opts.Metrics.ExecutionTime(name, func() { err = callFunc(ctx, pool) })

	return err
}

// TxContext the basic execution of a query in a transaction
func (s *_stmt) TxContext(name string, ctx context.Context, callFunc func(context.Context, *sql.Tx) error) error {
	return s.CallContext(name, ctx, func(ctx context.Context, db *sql.DB) error {
		dbx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		err = callFunc(ctx, dbx)
		if err != nil {
			return errors.Wrap(
				errors.Wrapf(err, "execute tx"),
				errors.Wrapf(dbx.Rollback(), "rollback tx"),
			)
		}

		return dbx.Commit()
	})
}
