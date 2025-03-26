/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

type jsonb struct {
	Any any
}

func (jb *jsonb) Scan(value any) error {
	if jb.Any == nil || value == nil {
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("type assertion to jsonb failed, got %T", value)
	}

	return json.Unmarshal(b, jb.Any)
}

func (jb *jsonb) Value() (driver.Value, error) {
	if jb.Any == nil {
		return nil, nil
	}

	b, err := json.Marshal(jb.Any)
	if err != nil {
		return nil, err
	}

	return driver.Value(b), nil
}

func pgCastTypes(args []any) []any {
	out := make([]any, 0, len(args))
	for _, arg := range args {
		out = append(out, pgCastType(arg))
	}
	return out
}

func pgCastType(arg any) any {
	switch a := arg.(type) {
	case json.Marshaler, json.Unmarshaler:
		return &jsonb{Any: a}

	case []bool:
		return (*pq.BoolArray)(&a)
	case []float64:
		return (*pq.Float64Array)(&a)
	case []float32:
		return (*pq.Float32Array)(&a)
	case []int64:
		return (*pq.Int64Array)(&a)
	case []int32:
		return (*pq.Int32Array)(&a)
	case []string:
		return (*pq.StringArray)(&a)
	case [][]byte:
		return (*pq.ByteaArray)(&a)

	case *[]bool:
		return (*pq.BoolArray)(a)
	case *[]float64:
		return (*pq.Float64Array)(a)
	case *[]float32:
		return (*pq.Float32Array)(a)
	case *[]int64:
		return (*pq.Int64Array)(a)
	case *[]int32:
		return (*pq.Int32Array)(a)
	case *[]string:
		return (*pq.StringArray)(a)
	case *[][]byte:
		return (*pq.ByteaArray)(a)
	}

	return arg
}
