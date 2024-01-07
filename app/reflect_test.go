/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestUnit_getReflectAddress(t *testing.T) {
	type (
		aa string
		bb struct{}
		ff func(_ string) bool
	)
	var (
		a    = 0
		b    = "0"
		c    = false
		d    = aa("aaa")
		e ff = func(_ string) bool { return false }
		f    = func(_ string) bool { return false }
		g    = errors.New("")
		h    = []string{}
		j    = bb{}
		k    = struct{}{}
	)

	tests := []struct {
		name string
		args reflect.Type
		obj  interface{}
		want string
		ok   bool
	}{
		{name: "Case1", args: reflect.TypeOf(a), obj: a, want: "int"},
		{name: "Case2", args: reflect.TypeOf(b), obj: b, want: "string"},
		{name: "Case3", args: reflect.TypeOf(c), obj: c, want: "bool"},
		{name: "Case4", args: reflect.TypeOf(d), obj: d, want: "go.osspkg.com/goppy/app.aa", ok: true},
		{name: "Case5", args: reflect.TypeOf(e), obj: e, want: "go.osspkg.com/goppy/app.ff", ok: true},
		{name: "Case6", args: reflect.TypeOf(f), obj: f, want: "func(string) bool", ok: true},
		{name: "Case6", args: reflect.TypeOf(f), obj: nil, want: "func(string) bool", ok: false},
		{name: "Case7", args: reflect.TypeOf(g), obj: g, want: "error"},
		{name: "Case8", args: reflect.TypeOf(h), obj: h, want: "[]string"},
		{name: "Case9", args: reflect.TypeOf(j), obj: j, want: "go.osspkg.com/goppy/app.bb", ok: true},
		{name: "Case10", args: reflect.TypeOf(k), obj: k, want: "struct {}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getReflectAddress(tt.args, tt.obj)
			if !strings.Contains(got, tt.want) {
				t.Errorf("getReflectAddress() = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("getReflectAddress() = %v, want %v", ok, tt.ok)
			}
		})
	}
}
