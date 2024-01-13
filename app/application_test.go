/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app_test

import (
	"testing"

	"go.osspkg.com/goppy/app"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_AppInvoke(t *testing.T) {
	out := ""
	call1 := func(s *Struct1) {
		s.Do(&out)
		out += "Done"
	}
	app.New().Modules(
		&Struct1{}, &Struct2{},
	).Invoke(call1)
	xtest.Equal(t, "[Struct1.Do]Done", out)

	out = ""
	call1 = func(s *Struct1) {
		s.Do2(&out)
		out += "Done"
	}
	app.New().ExitFunc(func(code int) {
		t.Log("Exit Code", code)
		xtest.Equal(t, 0, code)
	}).Modules(
		NewStruct1, &Struct2{},
	).Invoke(call1)
	xtest.Equal(t, "[Struct1.Do][Struct2.Do]Done", out)
}

type Struct1 struct{ s *Struct2 }

func NewStruct1(s2 *Struct2) *Struct1 {
	return &Struct1{s: s2}
}
func (*Struct1) Do(v *string) { *v += "[Struct1.Do]" }
func (s *Struct1) Do2(v *string) {
	*v += "[Struct1.Do]"
	s.s.Do(v)
}

type Struct2 struct{}

func (*Struct2) Do(v *string) { *v += "[Struct2.Do]" }
