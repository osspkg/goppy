/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iosync

import "sync"

type (
	Group interface {
		Wait()
		Background(call func())
		Run(call func())
	}

	_group struct {
		wg   sync.WaitGroup
		sync Switch
	}
)

func NewGroup() Group {
	return &_group{
		sync: NewSwitch(),
	}
}

func (v *_group) Wait() {
	v.sync.On()
	v.wg.Wait()
	v.sync.Off()
}

func (v *_group) Background(call func()) {
	if v.sync.IsOn() {
		return
	}
	v.wg.Add(1)
	go func() {
		call()
		v.wg.Done()
	}()
}

func (v *_group) Run(call func()) {
	if v.sync.IsOn() {
		return
	}
	v.wg.Add(1)
	call()
	v.wg.Done()
}
