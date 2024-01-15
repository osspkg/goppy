/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import (
	"fmt"
	"os"
	"reflect"
)

type AllowedKind struct {
	kind       []reflect.Kind
	typed      []reflect.Kind
	errMessage string
}

var (
	AllowedKindConfig = AllowedKind{
		kind:       []reflect.Kind{reflect.Ptr},
		errMessage: "Plugin.Config can only be a reference to an object",
	}
	AllowedKindInject = AllowedKind{
		kind: []reflect.Kind{reflect.Ptr, reflect.Func},
		typed: []reflect.Kind{
			reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String, reflect.Struct,
		},
		errMessage: "Plugin.Inject unsupported",
	}
	AllowedKindResolve = AllowedKind{
		kind:       []reflect.Kind{reflect.Func},
		errMessage: "Plugin.Resolve can only be a function that accepts dependencies",
	}
)

func (v AllowedKind) MustValidate(in interface{}) {
	into := reflect.TypeOf(in)
	for _, k := range v.kind {
		if into.Kind() == k {
			return
		}
	}
	if v.typed != nil {
		for _, k := range v.typed {
			if into.Kind() == k && into.Name() != k.String() {
				return
			}
		}
	}
	fmt.Printf("%s, but got `%T`\n", v.errMessage, in)
	os.Exit(1)
}
