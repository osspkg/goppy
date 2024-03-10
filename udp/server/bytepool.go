/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import "sync"

const bodyMaxSize = 65535

var bytesPool = sync.Pool{New: func() interface{} { return make([]byte, bodyMaxSize) }}

func getBuf() []byte {
	buf, ok := bytesPool.Get().([]byte)
	if !ok {
		buf = make([]byte, bodyMaxSize)
	}
	return buf
}

func setBuf(b []byte) {
	if len(b) != bodyMaxSize {
		return
	}
	bytesPool.Put(b) // nolint: staticcheck
}
