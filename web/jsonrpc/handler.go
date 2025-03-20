/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/web"
)

var (
	batchChar = byte('{')

	errFailReceiver = errors.New("receiver must be: func(context.Context, json.Unmarshaler) (json.Marshaler, error)")

	typeError   = reflect.TypeOf(new(error)).Elem()
	typeContext = reflect.TypeOf(new(context.Context)).Elem()
	typeJson    = reflect.TypeOf(new(Jsoner)).Elem()
)

type handle struct {
	model   reflect.Type
	handler reflect.Value
}

type Caller struct {
	routes *syncing.Map[string, handle]
}

func NewCaller() *Caller {
	return &Caller{
		routes: syncing.NewMap[string, handle](10),
	}
}

func (c *Caller) Add(method string, rcv any) error {
	rv := reflect.ValueOf(rcv)
	if rv.Kind() != reflect.Func {
		return errFailReceiver
	}

	rt := rv.Type()
	if rt.NumIn() != 2 || rt.NumOut() != 2 {
		return errFailReceiver
	}

	if !rt.In(0).Implements(typeContext) ||
		!rt.In(1).Implements(typeJson) ||
		!rt.Out(0).Implements(typeJson) ||
		!rt.Out(1).Implements(typeError) {
		return errFailReceiver
	}

	h := handle{
		model:   rt.In(1).Elem(),
		handler: rv,
	}

	c.routes.Set(strings.ToLower(method), h)
	return nil
}

func (c *Caller) Del(method string) {
	c.routes.Del(method)
}

func (c *Caller) Handler(ctx web.Context) {
	ctx.Header().Set("Content-Type", "application/json; charset=utf-8")

	var requests RequestBatch
	var single bool

	buf, err := ctx.BindRaw()
	if err != nil {
		logx.Error("Read JSONRPC request", "err", err)
		ctx.Error(400, ErrInvalidRequest)
		return
	}

	if buf.Len() == 0 {
		ctx.Error(400, ErrInvalidRequest)
		return
	}

	if fc := buf.Next(1); len(fc) > 0 && fc[0] == batchChar {
		single = true
	}

	if single {
		var req Request
		err = json.Unmarshal(buf.Bytes(), &req)
		if err == nil {
			requests = append(requests, req)
		}
	} else {
		err = json.Unmarshal(buf.Bytes(), &requests)
	}

	if err != nil {
		logx.Error("Decode JSONRPC request", "err", err)
		ctx.Error(400, ErrParseJSON)
		return
	}

	var wg sync.WaitGroup
	responses := make(ResponseBatch, 0, len(requests))

	wg.Add(len(requests))
	for _, req := range requests {
		h, ok := c.routes.Get(strings.ToLower(req.Method))
		if !ok {
			responses = append(responses, newErrorResponse(ErrMethodNotFound, req))
			wg.Done()
			continue
		}

		go func() {
			defer wg.Done()
			if re := do.Recovery(func() {
				model := reflect.New(h.model).Interface()

				if e0 := json.Unmarshal([]byte(req.Params), model); e0 != nil {
					responses = append(responses, newErrorResponse(e0, req))
					return
				}

				args := []reflect.Value{reflect.ValueOf(ctx.Context()), reflect.ValueOf(model)}
				result := h.handler.Call(args)

				if e1 := result[1].Interface(); e1 != nil {
					responses = append(responses, newErrorResponse(e1.(error), req)) //nolint:errcheck
					return
				}

				b, e2 := json.Marshal(result[0].Interface())
				if e2 != nil {
					responses = append(responses, newErrorResponse(e2, req)) //nolint:errcheck
					return
				}

				responses = append(responses, newResponse(b, req))
			}); re != nil {
				logx.Error("Call JSONRPC method", "method", req.Method, "err", re)
				responses = append(responses, newErrorResponse(ErrInternal, req))
			}
		}()
	}

	wg.Wait()

	var b []byte

	if single {
		if len(responses) > 0 {
			b, err = json.Marshal(responses[0])
		}
	} else {
		b, err = json.Marshal(responses)
	}

	if err != nil {
		logx.Error("Encode JSONRPC response", "err", err)

		if single {
			b, err = json.Marshal(newErrorResponse(ErrInternal, requests[0]))
		} else {
			b, err = json.Marshal(newErrorResponseBatch(ErrInternal, requests))
		}

		if err != nil {
			logx.Error("Encode JSONRPC errors", "err", err)
		}
	}

	ctx.Bytes(200, b)
}
