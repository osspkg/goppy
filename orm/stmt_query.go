/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"database/sql"
	"sync"
)

var poolQuery = sync.Pool{New: func() interface{} { return &query{} }}

type query struct {
	Q string
	P []interface{}
	B func(bind Scanner) error
}

func (v *query) SQL(query string, args ...interface{}) {
	v.Q, v.P = query, args
}

func (v *query) Bind(call func(bind Scanner) error) {
	v.B = call
}

func (v *query) Reset() *query {
	v.Q, v.P, v.B = "", nil, nil
	return v
}

type (
	// Scanner interface for bind data
	Scanner interface {
		Scan(args ...interface{}) error
	}

	// Querier interface for generate query
	Querier interface {
		SQL(query string, args ...interface{})
		Bind(call func(bind Scanner) error)
	}
)

// QueryContext ...
func (s *_stmt) QueryContext(name string, ctx context.Context, call func(q Querier)) error {
	return s.CallContext(name, ctx, func(ctx context.Context, db *sql.DB) error {
		return callQueryContext(ctx, db, call)
	})
}

func callQueryContext(ctx context.Context, db dbGetter, call func(q Querier)) error {
	q, ok := poolQuery.Get().(*query)
	if !ok {
		return errInvalidModelPool
	}
	defer poolQuery.Put(q.Reset())

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
