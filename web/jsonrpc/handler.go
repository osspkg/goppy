/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go.osspkg.com/do"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/web"
)

type Handler func(ctx context.Context, params json.RawMessage) (json.Marshaler, error)

type Caller struct {
	routes *syncing.Map[string, Handler]
}

func NewCaller() *Caller {
	return &Caller{
		routes: syncing.NewMap[string, Handler](10),
	}
}

func (c *Caller) Add(method string, handler Handler) {
	c.routes.Set(method, handler)
}

func (c *Caller) Del(method string) {
	c.routes.Del(method)
}

func (c *Caller) Handler(ctx web.Context) {
	ctx.Header().Set("Content-Type", "application/json; charset=utf-8")

	var reqs RequestBatch

	if err := ctx.BindJSON(&reqs); err != nil {
		logx.Error("Decode JSONRPC request", "err", err)

		b, err0 := json.Marshal(createErrorResponseBatch(err, reqs))
		if err0 != nil {
			logx.Error("Encode JSONRPC errors", "err", err0)
		}

		ctx.Bytes(200, b)
		return
	}

	var wg sync.WaitGroup
	resps := make(ResponseBatch, 0, len(reqs))

	for _, req := range reqs {
		handler, ok := c.routes.Get(req.Method)
		if !ok {
			resps = append(resps, newErrorResponse(fmt.Errorf("method not found: %s", req.Method), req))
			continue
		}

		do.Async(
			func() {
				defer wg.Done()

				resp, err := handler(ctx.Context(), req.Params)
				if err != nil {
					resps = append(resps, newErrorResponse(err, req))
				} else {
					resps = append(resps, newResponse(resp, req))
				}
			},
			func(err error) {
				logx.Error("Call JSONRPC method", "method", req.Method, "err", err)
			},
		)
	}

	wg.Wait()

	b, err := json.Marshal(resps)
	if err != nil {
		logx.Error("Encode JSONRPC response", "err", err)

		b, err = json.Marshal(createErrorResponseBatch(err, reqs))
		if err != nil {
			logx.Error("Encode JSONRPC errors", "err", err)
		}
	}

	ctx.Bytes(200, b)
}
