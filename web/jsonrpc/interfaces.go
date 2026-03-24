/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"encoding/json"

	"go.osspkg.com/goppy/v3/web"
)

type TError interface {
	GetCode() int64
	GetMessage() string
	GetContext() map[string]string
}

type THandleFunc func(ctx context.Context, wc web.Ctx, p json.RawMessage) (any, error)

type TApi interface {
	JSONRPCApiHandlers() map[string]THandleFunc
	RouteTags() []string
}
