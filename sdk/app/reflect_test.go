/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"reflect"
	"strings"
	"testing"

	"go.osspkg.com/goppy/sdk/errors"
)

func TestUnit_getRefAddr(t *testing.T) {
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
		want string
		ok   bool
	}{
		{name: "Case1", args: reflect.TypeOf(a), want: "int"},
		{name: "Case2", args: reflect.TypeOf(b), want: "string"},
		{name: "Case3", args: reflect.TypeOf(c), want: "bool"},
		{name: "Case4", args: reflect.TypeOf(d), want: "go.osspkg.com/goppy/sdk/app.aa", ok: true},
		{name: "Case5", args: reflect.TypeOf(e), want: "go.osspkg.com/goppy/sdk/app.ff", ok: true},
		{name: "Case6", args: reflect.TypeOf(f), want: "func(string) bool", ok: true},
		{name: "Case7", args: reflect.TypeOf(g), want: "error"},
		{name: "Case8", args: reflect.TypeOf(h), want: "[]string"},
		{name: "Case9", args: reflect.TypeOf(j), want: "go.osspkg.com/goppy/sdk/app.bb", ok: true},
		{name: "Case10", args: reflect.TypeOf(k), want: "struct {}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := getRefAddr(tt.args)
			if !strings.Contains(got, tt.want) {
				t.Errorf("getRefAddr() = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("getRefAddr() = %v, want %v", ok, tt.ok)
			}
		})
	}
}
