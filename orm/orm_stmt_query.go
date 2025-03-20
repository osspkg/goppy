/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"

	"go.osspkg.com/ioutils/pool"
)

var poolQuery = pool.New[*query](func() *query { return &query{} })

type query struct {
	Q string
	P []any
	B func(bind Scanner) error
}

func (v *query) SQL(query string, args ...any) {
	v.Q, v.P = query, args
}

func (v *query) Bind(call func(bind Scanner) error) {
	v.B = call
}

func (v *query) Reset() {
	v.Q, v.P, v.B = "", v.P[:0], nil
}

type (
	// Scanner interface for bind data
	Scanner interface {
		Scan(args ...any) error
	}

	// Querier interface for generate query
	Querier interface {
		SQL(query string, args ...any)
		Bind(call func(bind Scanner) error)
	}
)

func (v *_stmt) Query(ctx context.Context, name string, call func(q Querier)) error {
	return v.CallContext(ctx, name, func(ctx context.Context, db DB) error {
		return callQueryContext(ctx, db, call)
	})
}

func callQueryContext(ctx context.Context, db dbGetter, call func(q Querier)) error {
	q := poolQuery.Get()
	defer poolQuery.Put(q)

	call(q)

	rows, err := db.QueryContext(ctx, q.Q, q.P...)
	if err != nil {
		return err
	}
	defer rows.Close() // nolint: errcheck
	if q.B != nil {
		for rows.Next() {
			if err = q.B(rows); err != nil {
				return err
			}
		}
	}
	if err = rows.Close(); err != nil {
		return err
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return nil
}
