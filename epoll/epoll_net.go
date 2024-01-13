/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package epoll

import (
	"net"
	"sync"
)

type (
	epollNetMap   map[int]*epollNetItem
	epollNetSlice []*epollNetItem
)

type epollNetItem struct {
	Conn  net.Conn
	await bool
	Fd    int
	mux   sync.RWMutex
}

func (v *epollNetItem) Await(b bool) {
	v.mux.Lock()
	v.await = b
	v.mux.Unlock()
}

func (v *epollNetItem) IsAwait() bool {
	v.mux.RLock()
	is := v.await
	v.mux.RUnlock()
	return is
}
