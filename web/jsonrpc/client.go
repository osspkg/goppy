/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"go.osspkg.com/goppy/v2/web"
)

type Client struct {
	endpoint string
	cli      *web.ClientHttp
}

func NewClient(endpoint string, opts ...web.ClientHttpOption) *Client {
	return &Client{
		endpoint: endpoint,
		cli:      web.NewClientHttp(opts...),
	}
}

func (c *Client) Call(ctx context.Context, args ...Batch) error {
	var responses ResponseBatch
	indexes := make(map[string]int, len(args))
	requests := make(RequestBatch, 0, len(args))

	for i, arg := range args {
		req := Request{ID: uuid.New().String(), Method: arg.Method}

		b, err := arg.Params.MarshalJSON()
		if err != nil {
			return err
		}

		req.Params = b
		indexes[req.ID] = i
		requests = append(requests, req)
	}

	err := c.cli.Call(ctx, http.MethodPost, c.endpoint, requests, &responses)
	if err != nil {
		return err
	}

	for _, response := range responses {
		if i, ok := indexes[response.ID]; ok {
			args[i].Fallback(response.Result, response.Error)
		}
	}

	return nil
}
