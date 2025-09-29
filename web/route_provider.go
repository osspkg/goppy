/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

//go:generate easyjson

import (
	"net/http"
	"strings"

	"go.osspkg.com/xc"
)

type (
	route struct {
		name   string
		route  *BaseRouter
		serv   *Server
		config Config
	}

	// Router router handler interface
	Router interface {
		Use(args ...Middleware)
		NotFoundHandler(call func(ctx Ctx))
		RouteCollector
	}

	// RouteCollector interface of the router collection
	RouteCollector interface {
		Get(path string, call func(ctx Ctx))
		Head(path string, call func(ctx Ctx))
		Post(path string, call func(ctx Ctx))
		Put(path string, call func(ctx Ctx))
		Delete(path string, call func(ctx Ctx))
		Options(path string, call func(ctx Ctx))
		Patch(path string, call func(ctx Ctx))
		Match(path string, call func(ctx Ctx), methods ...string)
		Collection(prefix string, args ...Middleware) RouteCollector
	}
)

func newRouter(name string, c Config) *route {
	return &route{
		name:   name,
		route:  NewBaseRouter(),
		config: c,
	}
}

func (v *route) Up(c xc.Context) error {
	v.serv = NewServer(c.Context(), v.config, v.route)
	return v.serv.Up(c)
}
func (v *route) Down() error {
	return v.serv.Down()
}

func (v *route) Use(args ...Middleware) {
	for _, arg := range args {
		arg := arg
		v.route.Global(arg)
	}
}

func (v *route) NotFoundHandler(call func(ctx Ctx)) {
	v.route.NoFoundHandler(func(ctx Ctx) {
		call(ctx)
	})
}

func (v *route) Match(path string, call func(ctx Ctx), methods ...string) {
	v.route.Route(path, func(ctx Ctx) {
		call(ctx)
	}, methods...)
}

func (v *route) Get(path string, call func(ctx Ctx))     { v.Match(path, call, http.MethodGet) }
func (v *route) Head(path string, call func(ctx Ctx))    { v.Match(path, call, http.MethodHead) }
func (v *route) Post(path string, call func(ctx Ctx))    { v.Match(path, call, http.MethodPost) }
func (v *route) Put(path string, call func(ctx Ctx))     { v.Match(path, call, http.MethodPut) }
func (v *route) Delete(path string, call func(ctx Ctx))  { v.Match(path, call, http.MethodDelete) }
func (v *route) Options(path string, call func(ctx Ctx)) { v.Match(path, call, http.MethodOptions) }
func (v *route) Patch(path string, call func(ctx Ctx))   { v.Match(path, call, http.MethodPatch) }

// Collection route collection handler
func (v *route) Collection(prefix string, args ...Middleware) RouteCollector {
	prefix = "/" + strings.Trim(prefix, "/")
	for _, arg := range args {
		arg := arg
		v.route.Middlewares(prefix, arg)
	}
	return &rc{
		p: prefix,
		r: v,
	}
}

type rc struct {
	p string
	r *route
}

func (v *rc) Match(path string, call func(ctx Ctx), methods ...string) {
	path = strings.TrimRight(v.p, "/") + "/" + strings.Trim(path, "/")
	v.r.Match(path, call, methods...)
}

func (v *rc) Get(path string, call func(ctx Ctx))     { v.Match(path, call, http.MethodGet) }
func (v *rc) Head(path string, call func(ctx Ctx))    { v.Match(path, call, http.MethodHead) }
func (v *rc) Post(path string, call func(ctx Ctx))    { v.Match(path, call, http.MethodPost) }
func (v *rc) Put(path string, call func(ctx Ctx))     { v.Match(path, call, http.MethodPut) }
func (v *rc) Delete(path string, call func(ctx Ctx))  { v.Match(path, call, http.MethodDelete) }
func (v *rc) Options(path string, call func(ctx Ctx)) { v.Match(path, call, http.MethodOptions) }
func (v *rc) Patch(path string, call func(ctx Ctx))   { v.Match(path, call, http.MethodPatch) }

func (v *rc) Collection(prefix string, args ...Middleware) RouteCollector {
	v.p = strings.TrimRight(v.p, "/") + "/" + strings.Trim(prefix, "/")
	for _, arg := range args {
		arg := arg
		v.r.route.Middlewares(v.p, arg)
	}
	return v
}
