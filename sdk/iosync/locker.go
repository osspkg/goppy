/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iosync

import "sync"

type (
	Lock interface {
		RLock(call func())
		Lock(call func())
	}
	_lock struct {
		mux sync.RWMutex
	}
)

func NewLock() Lock {
	return &_lock{}
}

func (v *_lock) Lock(call func()) {
	v.mux.Lock()
	call()
	v.mux.Unlock()
}
func (v *_lock) RLock(call func()) {
	v.mux.RLock()
	call()
	v.mux.RUnlock()
}
