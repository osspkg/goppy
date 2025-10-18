/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"

	"go.osspkg.com/errors"

	"go.osspkg.com/goppy/v2/orm/metric"
)

// PingContext database ping
func (v *_stmt) PingContext(ctx context.Context) (err error) {
	if err = v.err.Get(); err != nil {
		return
	}

	metric.ExecTime(v.tag, "ping", func() {
		if err = v.db.PingContext(ctx); err != nil {
			err = fmt.Errorf("failed pinging database: %w", err)
		}
	})

	return
}

// CallContext basic query execution
func (v *_stmt) CallContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) (err error) {
	if err = v.err.Get(); err != nil {
		return
	}

	metric.ExecTime(v.tag, name, func() {
		if err = callFunc(ctx, v.db); err != nil {
			err = fmt.Errorf("failed calling %s: %w", name, err)
		}
	})

	return
}

// TxContext the basic execution of a query in a transaction
func (v *_stmt) TxContext(ctx context.Context, name string, callFunc func(context.Context, DB) error) error {
	if err := v.err.Get(); err != nil {
		return err
	}

	dbx, err := v.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed beginning transaction: %w", err)
	}

	metric.ExecTime(v.tag, name, func() {
		if err = callFunc(ctx, dbx); err != nil {
			err = fmt.Errorf("failed calling %s: %w", name, err)
		}
	})

	if err != nil {
		return errors.Wrap(
			errors.Wrapf(err, "failed execute transaction"),
			errors.Wrapf(dbx.Rollback(), "failed rollback transaction"),
		)
	}

	if err = dbx.Commit(); err != nil {
		return fmt.Errorf("failed committing transaction: %w", err)
	}

	return nil
}
