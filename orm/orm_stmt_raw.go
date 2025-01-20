/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"

	"go.osspkg.com/errors"
)

// PingContext database ping
func (v *_stmt) PingContext(ctx context.Context) (err error) {
	if v.err != nil {
		return v.err
	}
	execTime(v.tag, "ping", func() {
		err = v.conn.PingContext(ctx)
	})
	return
}

// CallContext basic query execution
func (v *_stmt) CallContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) (err error) {
	if v.err != nil {
		return v.err
	}
	execTime(v.tag, name, func() {
		err = callFunc(ctx, v.conn)
	})
	return
}

// TxContext the basic execution of a query in a transaction
func (v *_stmt) TxContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) error {
	if v.err != nil {
		return v.err
	}

	dbx, err := v.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	execTime(v.tag, name, func() {
		err = callFunc(ctx, dbx)
	})
	if err != nil {
		return errors.Wrap(
			errors.Wrapf(err, "execute tx"),
			errors.Wrapf(dbx.Rollback(), "rollback tx"),
		)
	}

	return dbx.Commit()
}
