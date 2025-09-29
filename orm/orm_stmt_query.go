/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"context"
	"fmt"

	"go.osspkg.com/ioutils/pool"

	"go.osspkg.com/goppy/v2/orm/dialect"
)

var poolQuery = pool.New[*query](func() *query { return &query{} })

type query struct {
	Dialect dialect.Connector
	Query   string
	Args    []any
	Binding func(bind Scanner) error
}

func (v *query) SQL(query string, args ...any) {
	if cast := v.Dialect.CastTypesFunc(); cast != nil {
		cast(args)
	}

	v.Query, v.Args = query, args
}

func (v *query) Bind(call func(bind Scanner) error) {
	v.Binding = call
}

func (v *query) Reset() {
	v.Query, v.Args, v.Binding, v.Dialect = "", v.Args[:0], nil, nil
}

type scan struct {
	Dialect  dialect.Connector
	Scanning Scanner
}

func (v *scan) Scan(args ...any) error {
	if cast := v.Dialect.CastTypesFunc(); cast != nil {
		cast(args)
	}

	if err := v.Scanning.Scan(args...); err != nil {
		return fmt.Errorf("failed scan: %w", err)
	}

	return nil
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

func callQueryContext(ctx context.Context, db queryGetter, call func(q Querier), dc dialect.Connector) error {
	obj := poolQuery.Get()
	defer func() { poolQuery.Put(obj) }()

	obj.Dialect = dc

	call(obj)

	rows, err := db.QueryContext(ctx, obj.Query, obj.Args...)
	if err != nil {
		return err
	}
	defer rows.Close() // nolint: errcheck

	if obj.Binding != nil {
		for rows.Next() {
			if err = obj.Binding(&scan{Dialect: dc, Scanning: rows}); err != nil {
				return fmt.Errorf("failed binding: %w", err)
			}
		}
	}

	if err = rows.Close(); err != nil {
		return fmt.Errorf("failed closing rows: %w", err)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("failed closing rows: %w", err)
	}

	return nil
}
