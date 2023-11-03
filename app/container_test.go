/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"fmt"
	"testing"

	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xtest"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type t0 struct{}

func newT0() *t0           { return &t0{} }
func (t0 *t0) Up() error   { return nil }
func (t0 *t0) Down() error { return nil }
func (t0 *t0) V() string   { return "t0V" }

type t1 struct {
	t0 *t0
}

func newT1(t0 *t0) *t1     { return &t1{t0: t0} }
func (t1 *t1) Up() error   { return nil }
func (t1 *t1) Down() error { return nil }
func (t1 *t1) V() string   { return "t1V" }

type t2 struct {
	t0 *t0
	t1 *t1
}

func newT2(t1 *t1, t0 *t0) *t2             { return &t2{t0: t0, t1: t1} }
func (t2 *t2) Up() error                   { return nil }
func (t2 *t2) Down() error                 { return nil }
func (t2 *t2) V() (string, string, string) { return "t2V", t2.t1.V(), t2.t0.V() }

type t4 struct {
	T0  *t0
	T1  *t1
	T2  *t2
	T7  *t7
	T44 t44
}

type t44 struct {
	Env string
}

type t5 struct{}

func newT5() *t5         { return &t5{} }
func (t5 *t5) V() string { return "t5V" }

type t6 struct{ T4 t4 }

func newT6(t4 t4) *t6    { return &t6{T4: t4} }
func (t6 *t6) V() string { return "t6V" }

type t7 struct{}

func newT7() *t7         { return &t7{} }
func (t7 *t7) V() string { return "t7V" }

type t8 struct{}

func newT8() (*t8, error) { return &t8{}, nil }
func (t8 *t8) V() string  { return "t8V" }

type hello string

var AA = hello("hhhh")

type ii interface {
	V() string
}

func newT7i(_ hello) ii {
	return &t7{}
}

func TestUnit_Dependencies(t *testing.T) {
	dep := newDic(xc.New())

	xtest.NoError(t, dep.Register([]interface{}{
		newT1, newT2, newT5, newT6, newT7(), newT8,
		AA, newT7i, newT0, t44{Env: "aaa"},
		func(b *t6) {
			t.Log("anonymous function")
		},
	}...))

	xtest.NoError(t, dep.Build())
	xtest.Error(t, dep.Build())

	xtest.NoError(t, dep.Inject(func(
		v1 *t1, v2 *t2, v3 *t5, v4 *t6, v5 *t6,
		v6 *t7, v7 *t8, v8 hello,
		v9 ii, v10 *t0, v11 t44,
	) {
		xtest.Equal(t, "t1V", v1.V())
		vv2, _, _ := v2.V()
		xtest.Equal(t, "t2V", vv2)
		xtest.Equal(t, "t5V", v3.V())
		xtest.Equal(t, "t6V", v4.V())
		xtest.Equal(t, "t6V", v5.V())
		xtest.Equal(t, "t7V", v6.V())
		xtest.Equal(t, hello("hhhh"), v8)
		xtest.Equal(t, "t7V", v9.V())
		xtest.Equal(t, "t0V", v10.V())
		xtest.Equal(t, "aaa", v11.Env)
	}))

	xtest.Error(t, dep.Inject(func(a string, b int, c bool) {

	}))

	xtest.NoError(t, dep.Down())
	xtest.Error(t, dep.Down())
}

type demo1 struct{}
type demo2 struct{}
type demo3 struct{}

func newDemo() (*demo1, *demo2, *demo3) { return &demo1{}, &demo2{}, &demo3{} }
func (d *demo1) Up() error {
	fmt.Println("demo1 up")
	return nil
}
func (d *demo1) Down() error {
	fmt.Println("demo1 down")
	return nil
}

func TestUnit_Dependencies2(t *testing.T) {
	dep := newDic(xc.New())
	xtest.NoError(t, dep.Register([]interface{}{
		newDemo,
	}...))
	xtest.NoError(t, dep.Build())
	xtest.Error(t, dep.Build())
	xtest.NoError(t, dep.Down())
	xtest.Error(t, dep.Down())
}

type demo4 struct{}

func newDemo4() (*demo4, error) { return nil, fmt.Errorf("fail init constructor demo4") }

func TestUnit_Dependencies3(t *testing.T) {
	dep := newDic(xc.New())
	xtest.NoError(t, dep.Register([]interface{}{
		newDemo4,
	}...))
	err := dep.Build()
	xtest.Error(t, err)
	fmt.Println(err.Error())
	xtest.Contains(t, err.Error(), "fail init constructor demo4")
}

func newDemo5() error { return fmt.Errorf("fail init constructor newDemo5") }

func TestUnit_Dependencies4(t *testing.T) {
	dep := newDic(xc.New())
	xtest.NoError(t, dep.Register(newDemo5))
	err := dep.Build()
	xtest.Error(t, err)
	fmt.Println(err.Error())
	xtest.Contains(t, err.Error(), "fail init constructor newDemo5")
}

type demo6 struct{}

func newDemo6() *demo6 { return &demo6{} }
func (d *demo6) Up() error {
	fmt.Println("demo6 up")
	return nil
}
func (d *demo6) Down() error {
	fmt.Println("demo6 down")
	return nil
}
func (d *demo6) Name() string {
	return "DEMO 6"
}

func TestUnit_DicInvoke1(t *testing.T) {
	dep := newDic(xc.New())
	xtest.NoError(t, dep.Register(newDemo6))
	xtest.NoError(t, dep.Invoke(func(d *demo6) {
		fmt.Println("Invoke", d.Name())
	}))
	xtest.NoError(t, dep.Down())
	xtest.Error(t, dep.Down())
}

func TestUnit_DicInvoke2(t *testing.T) {
	dep := newDic(xc.New())
	xtest.NoError(t, dep.Register(newDemo6))
	xtest.NoError(t, dep.Invoke(func() {
		fmt.Println("Invoke")
	}))
	xtest.NoError(t, dep.Down())
	xtest.Error(t, dep.Down())
}
