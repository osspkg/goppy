/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"context"
	"strings"
	"time"

	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/web"
)

type Transport interface {
	Add(r TApi)
}

type service struct {
	opt      *options
	handlers map[string]*syncing.Map[string, THandleFunc]
	routes   web.ServerPool
}

func newService(routes web.ServerPool, opts ...Option) Transport {
	obj := &service{
		opt: &options{
			timeout: time.Second * 5,
			path:    "/jsonrpc",
			errHandler: func(_ string, err error) error {
				return err
			},
		},
		handlers: make(map[string]*syncing.Map[string, THandleFunc], 2),
		routes:   routes,
	}

	for _, o := range opts {
		o(obj.opt)
	}

	return obj
}

func (v *service) Down() error {
	for tag, r := range v.handlers {
		r.Reset()

		logx.Info("JSON-RPC Transport",
			"do", "stop",
			"tag", tag)
	}
	return nil
}

func (v *service) Up() error {
	v.routes.All(func(tag string, r web.Router) {
		resolve, ok := v.handlers[tag]
		if !ok {
			return
		}

		r.Post(v.opt.path, v.Handle(resolve))

		logx.Info("JSON-RPC Transport",
			"do", "start",
			"tag", tag)
	})

	return nil
}

func (v *service) Add(r TApi) {
	for _, tag := range r.RouteTags() {
		resolve, ok := v.handlers[tag]
		if !ok {
			resolve = syncing.NewMap[string, THandleFunc](10)
			v.handlers[tag] = resolve
		}

		for method, handler := range r.JSONRPCApiHandlers() {
			resolve.Set(method, handler)
		}
	}
}

func (v *service) Handle(resolve *syncing.Map[string, THandleFunc]) func(wc web.Ctx) {
	return func(wc web.Ctx) {
		req := poolRequestRaw.Get()
		defer poolRequestRaw.Put(req)

		if err := wc.BindJSON(req); err != nil {
			wc.String(400, v.opt.errHandler("", err).Error())
			return
		}

		res := poolResponseAnySync.Get()
		defer poolResponseAnySync.Put(res)

		ctx, cancel := context.WithTimeout(wc.Context(), v.opt.timeout)
		defer cancel()

		wg := syncing.NewGroup(ctx)
		wg.OnPanic(func(err error) {
			logx.Error("json-rpc handle panic", "err", err)
		})

		for _, item := range *req {
			item := item

			method := strings.ToLower(item.Method)

			wg.Background(method, func(ctx context.Context) {
				out := responseAny{
					Id: item.Id,
				}

				if handler, ok := resolve.Get(method); ok {

					result, err := handler(ctx, wc, item.Params)
					if err != nil {
						out.Error = errConvert(v.opt.errHandler(method, err))
					} else {
						out.Result = result
					}

				} else {
					out.Error = errConvert(ErrUnsupportedMethod)
				}

				res.Append(out)
			})

		}

		wg.Wait()
		wc.JSON(200, bulkResponseAny(res.Extract()))
	}
}
