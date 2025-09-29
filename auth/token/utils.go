/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"go.osspkg.com/ioutils/data"
	"go.osspkg.com/ioutils/pool"
)

var dataPool = pool.New[*data.Buffer](func() *data.Buffer {
	return data.NewBuffer(1024)
})
