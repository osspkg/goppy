/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package appreflect

import (
	"reflect"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/plugins"
)

func AnySlice(arg any) []any {
	refVal := reflect.ValueOf(arg)
	if refVal.Kind() != reflect.Slice {
		return []any{arg}
	}

	result := make([]any, 0, refVal.Len())
	for i := 0; i < refVal.Len(); i++ {
		result = append(result, refVal.Index(i).Interface())
	}

	return result
}

func Validate(arg any, k plugins.AllowedKind, call func(any) error) {
	if arg == nil {
		return
	}
	console.FatalIfErr(k.Validate(arg), "validate dependency")
	console.FatalIfErr(call(arg), "append dependency")
}
