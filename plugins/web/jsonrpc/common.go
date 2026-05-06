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
	ErrUnsupportedMethod = errors.New("unsupported method")
	ErrNoResponse        = errors.New("no response")
)

var (
	poolRequestRaw = pool.New[*bulkRequestRaw](func() *bulkRequestRaw {
		br := make(bulkRequestRaw, 0, 2)
		return &br
	})

	poolRequestAny = pool.New[*bulkRequestAny](func() *bulkRequestAny {
		br := make(bulkRequestAny, 0, 2)
		return &br
	})

	poolResponseRaw = pool.New[*bulkResponseRaw](func() *bulkResponseRaw {
		br := make(bulkResponseRaw, 0, 2)
		return &br
	})

	poolResponseAnySync = pool.New[*syncing.Slice[responseAny]](func() *syncing.Slice[responseAny] {
		return syncing.NewSlice[responseAny](uint(2))
	})
)
