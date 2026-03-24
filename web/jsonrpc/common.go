/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"errors"

	"go.osspkg.com/ioutils/pool"
	"go.osspkg.com/syncing"
)

var (
	poolResponse = pool.New[*syncing.Slice[response]](func() *syncing.Slice[response] {
		return syncing.NewSlice[response](uint(2))
	})

	poolRequest = pool.New[*bulkRequest](func() *bulkRequest {
		br := make(bulkRequest, 0, 2)
		return &br
	})

	ErrUnsupportedMethod = errors.New("unsupported method")
)
