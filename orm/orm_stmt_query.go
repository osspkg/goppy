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
	D string
	Q string
	P []any
	B func(bind Scanner) error
}

func (v *query) SQL(query string, args ...any) {
	switch v.D {
	case PgSQLDialect:
		applyPGSqlCastTypes(args)
	default:
	}

	v.Q, v.P = query, args
}

func (v *query) Bind(call func(bind Scanner) error) {
	v.B = call
}

func (v *query) Reset() {
	v.Q, v.P, v.B, v.D = "", v.P[:0], nil, ""
}

type scan struct {
	D string
	S Scanner
}

func (v *scan) Scan(args ...any) error {
	switch v.D {
	case PgSQLDialect:
		applyPGSqlCastTypes(args)
	default:
	}

	return v.S.Scan(args...)
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
		return callQueryContext(ctx, db, call, v.dialect)
	})
}

func callQueryContext(ctx context.Context, db dbGetter, call func(q Querier), dialect string) error {
	q := poolQuery.Get()
	defer func() { poolQuery.Put(q) }()

	q.D = dialect

	call(q)

	rows, err := db.QueryContext(ctx, q.Q, q.P...)
	if err != nil {
		return err
	}
	defer rows.Close() // nolint: errcheck
	if q.B != nil {
		for rows.Next() {
			if err = q.B(&scan{D: dialect, S: rows}); err != nil {
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
