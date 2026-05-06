/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package rpc

import "context"

type rpcPlugin interface {
	Start(ctx context.Context, opts map[string]string) error
	Stop() error
	Call(ctx context.Context, method string, params, result any) (err error)
}
