/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/lib/pq"
)

type (
	TypeScanValuer interface {
		driver.Valuer
		sql.Scanner
	}

	TypeJSONb interface {
		json.Marshaler
		json.Unmarshaler
	}
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

func applyPGSqlCastTypes(args []any) {
	count := len(args)

	for i := 0; i < count; i++ {
		if reflect.ValueOf(args[i]).Kind() == reflect.Slice {
			args[i] = pq.Array(args[i])
			continue
		}

		switch args[i].(type) {
		case TypeScanValuer:
			// ---
		case []byte, string, *[]byte, *string:
			// ---
		case int, int8, int16, int32, int64, *int, *int8, *int16, *int32, *int64:
			// ---
		case uint, uint8, uint16, uint32, uint64, *uint, *uint8, *uint16, *uint32, *uint64:
			// ---
		case float32, float64, *float32, *float64:
			// ---
		case bool, *bool:
			// ---
		case complex64, complex128, *complex64, *complex128:
			// ---
		case time.Time, *time.Time:
			// ---
		case TypeJSONb:
			args[i] = &jsonb{Any: args[i]}
		default:
		}
	}
}
