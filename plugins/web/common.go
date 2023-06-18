/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

type rwlocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

func lock(l rwlocker, call func()) {
	l.Lock()
	call()
	l.Unlock()
}
func rwlock(l rwlocker, call func()) {
	l.RLock()
	call()
	l.RUnlock()
}
