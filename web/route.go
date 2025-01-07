/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"net/http"
	"sync"
)

var _ http.Handler = (*BaseRouter)(nil)

type BaseRouter struct {
	handler *ctrlHandler
	lock    sync.RWMutex
}

func NewBaseRouter() *BaseRouter {
	return &BaseRouter{
		handler: newCtrlHandler(),
	}
}

// Route add new route
func (v *BaseRouter) Route(path string, ctrl func(http.ResponseWriter, *http.Request), methods ...string) {
	v.lock.Lock()
	v.handler.Route(path, ctrl, methods)
	v.lock.Unlock()
}

// Global add global middlewares
func (v *BaseRouter) Global(
	middlewares ...func(func(http.ResponseWriter, *http.Request),
	) func(http.ResponseWriter, *http.Request)) {
	v.lock.Lock()
	v.handler.Middlewares("", middlewares...)
	v.lock.Unlock()
}

// Middlewares add middlewares to route
func (v *BaseRouter) Middlewares(
	path string, middlewares ...func(func(http.ResponseWriter, *http.Request),
	) func(http.ResponseWriter, *http.Request)) {
	v.lock.Lock()
	v.handler.Middlewares(path, middlewares...)
	v.lock.Unlock()
}

// NoFoundHandler ctrlHandler call if route not found
func (v *BaseRouter) NoFoundHandler(call func(http.ResponseWriter, *http.Request)) {
	v.lock.Lock()
	v.handler.NoFoundHandler(call)
	v.lock.Unlock()
}

// ServeHTTP http interface
func (v *BaseRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.lock.RLock()
	defer v.lock.RUnlock()

	code, next, params, midd := v.handler.Match(r.URL.Path, r.Method)
	if code != http.StatusOK {
		next = codeHandler(code)
	}

	ctx := r.Context()
	for key, val := range params {
		ctx = context.WithValue(ctx, uriParamKey(key), val)
	}

	for i := len(midd) - 1; i >= 0; i-- {
		next = midd[i](next)
	}
	next(w, r.WithContext(ctx))
}

func codeHandler(code int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	}
}
