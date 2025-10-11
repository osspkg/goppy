/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package pgsql

import (
	"reflect"
	"time"

	"github.com/lib/pq"

	"go.osspkg.com/goppy/v2/orm/custom_type"
)

func (p *pool) CastTypesFunc() func(args []any) {
	return func(args []any) {
		count := len(args)
		for i := 0; i < count; i++ {

			switch args[i].(type) {
			case []byte, string, *[]byte, *string:
				continue
			case int, int8, int16, int32, int64, *int, *int8, *int16, *int32, *int64:
				continue
			case uint, uint8, uint16, uint32, uint64, *uint, *uint8, *uint16, *uint32, *uint64:
				continue
			case float32, float64, *float32, *float64:
				continue
			case bool, *bool:
				continue
			case complex64, complex128, *complex64, *complex128:
				continue
			case time.Time, *time.Time:
				continue
			case custom_type.ScanValuerInterface:
				continue
			case custom_type.JSONbInterface:
				args[i] = &custom_type.JSONb{Any: args[i]}
			default:
			}

			ref := reflect.ValueOf(args[i])

			if ref.Kind() == reflect.Slice || ref.Kind() == reflect.Array {
				args[i] = pq.Array(args[i])
				continue
			}

			if ref.Kind() == reflect.Ptr &&
				(ref.Elem().Kind() == reflect.Slice || ref.Elem().Kind() == reflect.Array) {
				args[i] = pq.Array(args[i])
				continue
			}
		}
	}
}
