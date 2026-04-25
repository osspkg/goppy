/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"go.osspkg.com/goppy/v3/web/client"
	"go.osspkg.com/goppy/v3/web/client/comparison"
)

type Client struct {
	address string
	cli     client.HTTPClient
	genID   func() string
}

func New(address string, opts ...Opt) *Client {
	cliOpts := &cliopts{
		genID: uuid.NewString,
		httpopts: []client.HTTPOption{
			client.WithComparisonType(
				comparison.JSON{Force: true},
			),
		},
	}

	for _, opt := range opts {
		opt(cliOpts)
	}

	return &Client{
		address: address,
		cli:     client.NewHTTPClient(cliOpts.httpopts...),
		genID:   cliOpts.genID,
	}
}

func (c *Client) Call(
	ctx context.Context, method string,
	params json.Marshaler, result json.Unmarshaler,
) error {
	ch := &Chunk{
		Method: method,
		Params: params,
		Result: result,
	}

	if err := c.BulkCall(ctx, ch); err != nil {
		return err
	}

	return ch.Error
}

type Chunk struct {
	Method string
	Params json.Marshaler
	Result json.Unmarshaler
	Error  error
}

func (c *Client) BulkCall(ctx context.Context, bulk ...*Chunk) error {
	if len(bulk) == 0 {
		return nil
	}

	req := poolRequestAny.Get()
	defer poolRequestAny.Put(req)
	res := poolResponseRaw.Get()
	defer poolResponseRaw.Put(res)

	ids := make(map[string]*Chunk, len(bulk))

	for _, ch := range bulk {
		id := c.genID()
		ids[id] = ch
		*req = append(*req, requestAny{
			Id:     id,
			Method: ch.Method,
			Params: ch.Params,
		})
	}

	if err := c.cli.Send(ctx, http.MethodPost, c.address, req, &res); err != nil {
		return err
	}

	for _, re := range *res {
		if ch, ok := ids[re.Id]; ok {
			delete(ids, re.Id)

			if err := re.Error; err != nil {
				ch.Error = err
				continue
			}

			if err := ch.Result.UnmarshalJSON(re.Result); err != nil {
				ch.Error = fmt.Errorf(
					"failed to unmarshal result: %w [requestId: %s]", err, re.Id)
			}
		}
	}

	for _, ch := range ids {
		ch.Error = ErrNoResponse
	}

	return nil
}

type ModelAdapter[T any] struct {
	Data T
}

func (m *ModelAdapter[T]) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.Data)
}

func (m ModelAdapter[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Data)
}
