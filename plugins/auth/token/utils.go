/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"go.osspkg.com/bb"
	"go.osspkg.com/ioutils/pool"
)

var dataPool = pool.New[*bb.Buffer](func() *bb.Buffer {
	return bb.New(1024)
})
