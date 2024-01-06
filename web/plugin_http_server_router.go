/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

//go:generate easyjson

import (
	"fmt"
	"net/http"
	"strings"

	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type (
	routePoolItem struct {
		active bool
		route  *route
	}

	//RouterPool router pool handler
	RouterPool interface {
		//All method to get all route handlers
		All(call func(name string, router Router))
		//Main method to get Main route handler
		Main() Router
		//Get method to get route handler by key
		Get(name string) Router
	}

	routeProvider struct {
		pool map[string]*routePoolItem
	}
)

func newRouteProvider(configs map[string]ConfigHttp, l xlog.Logger) *routeProvider {
	v := &routeProvider{
		pool: make(map[string]*routePoolItem),
	}
	for name, config := range configs {
		v.pool[name] = &routePoolItem{
			active: false,
			route:  newRouter(config, l),
		}
	}
	return v
}

// All method to get all route handlers
func (v *routeProvider) All(call func(name string, router Router)) {
	for n, r := range v.pool {
		call(n, r.route)
	}
}

// Main method to get Main route handler
func (v *routeProvider) Main() Router {
	return v.Get("main")
}

// Get method to get route handler by key
func (v *routeProvider) Get(name string) Router {
	if r, ok := v.pool[name]; ok {
		return r.route
	}
	panic(fmt.Sprintf("Route with name `%s` is not found", name))
}

func (v *routeProvider) Up(c xc.Context) error {
	for n, r := range v.pool {
		r.active = true
		if err := r.route.Up(c); err != nil {
			return fmt.Errorf("pool `%s`: %w", n, err)
		}
	}
	return nil
}

func (v *routeProvider) Down() error {
	for n, r := range v.pool {
		if !r.active {
			continue
		}
		if err := r.route.Down(); err != nil {
			return fmt.Errorf("pool `%s`: %w", n, err)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	route struct {
		route  *BaseRouter
		serv   *ServerHttp
		config ConfigHttp
		log    xlog.Logger
	}

	//Router router handler interface
	Router interface {
		Use(args ...Middleware)
		NotFoundHandler(call func(ctx Context))
		RouteCollector
	}

	//RouteCollector interface of the router collection
	RouteCollector interface {
		Get(path string, call func(ctx Context))
		Head(path string, call func(ctx Context))
		Post(path string, call func(ctx Context))
		Put(path string, call func(ctx Context))
		Delete(path string, call func(ctx Context))
		Options(path string, call func(ctx Context))
		Patch(path string, call func(ctx Context))
		Match(path string, call func(ctx Context), methods ...string)
		Collection(prefix string, args ...Middleware) RouteCollector
	}
)

func newRouter(c ConfigHttp, l xlog.Logger) *route {
	return &route{
		route:  NewBaseRouter(),
		config: c,
		log:    l,
	}
}

func (v *route) Up(c xc.Context) error {
	v.serv = NewServerHttp(v.config, v.route, v.log)
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

func (v *route) NotFoundHandler(call func(ctx Context)) {
	v.route.NoFoundHandler(func(w http.ResponseWriter, r *http.Request) {
		call(newContext(w, r, v.log))
	})
}

func (v *route) Match(path string, call func(ctx Context), methods ...string) {
	v.route.Route(path, func(w http.ResponseWriter, r *http.Request) {
		call(newContext(w, r, v.log))
	}, methods...)
}

func (v *route) Get(path string, call func(ctx Context))     { v.Match(path, call, http.MethodGet) }
func (v *route) Head(path string, call func(ctx Context))    { v.Match(path, call, http.MethodHead) }
func (v *route) Post(path string, call func(ctx Context))    { v.Match(path, call, http.MethodPost) }
func (v *route) Put(path string, call func(ctx Context))     { v.Match(path, call, http.MethodPut) }
func (v *route) Delete(path string, call func(ctx Context))  { v.Match(path, call, http.MethodDelete) }
func (v *route) Options(path string, call func(ctx Context)) { v.Match(path, call, http.MethodOptions) }
func (v *route) Patch(path string, call func(ctx Context))   { v.Match(path, call, http.MethodPatch) }

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

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type rc struct {
	p string
	r *route
}

func (v *rc) Match(path string, call func(ctx Context), methods ...string) {
	path = strings.TrimRight(v.p, "/") + "/" + strings.Trim(path, "/")
	v.r.Match(path, call, methods...)
}

func (v *rc) Get(path string, call func(ctx Context))     { v.Match(path, call, http.MethodGet) }
func (v *rc) Head(path string, call func(ctx Context))    { v.Match(path, call, http.MethodHead) }
func (v *rc) Post(path string, call func(ctx Context))    { v.Match(path, call, http.MethodPost) }
func (v *rc) Put(path string, call func(ctx Context))     { v.Match(path, call, http.MethodPut) }
func (v *rc) Delete(path string, call func(ctx Context))  { v.Match(path, call, http.MethodDelete) }
func (v *rc) Options(path string, call func(ctx Context)) { v.Match(path, call, http.MethodOptions) }
func (v *rc) Patch(path string, call func(ctx Context))   { v.Match(path, call, http.MethodPatch) }

func (v *rc) Collection(prefix string, args ...Middleware) RouteCollector {
	v.p = strings.TrimRight(v.p, "/") + "/" + strings.Trim(prefix, "/")
	for _, arg := range args {
		arg := arg
		v.r.route.Middlewares(v.p, arg)
	}
	return v
}
