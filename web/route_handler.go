/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"
	"strings"
)

const anyPath = "#"

type ctrlHandler struct {
	list        map[string]*ctrlHandler
	methods     map[string]func(ctx Ctx)
	matcher     *paramMatch
	middlewares []Middleware
	notFound    func(ctx Ctx)
}

func newCtrlHandler() *ctrlHandler {
	return &ctrlHandler{
		list:        make(map[string]*ctrlHandler),
		methods:     make(map[string]func(ctx Ctx)),
		matcher:     newParamMatch(),
		middlewares: make([]Middleware, 0),
	}
}

func (v *ctrlHandler) append(path string) *ctrlHandler {
	if uh, ok := v.list[path]; ok {
		return uh
	}
	uh := newCtrlHandler()
	v.list[path] = uh
	return uh
}

func (v *ctrlHandler) next(path string, vars uriParamData) (*ctrlHandler, bool) {
	if uh, ok := v.list[path]; ok {
		return uh, false
	}
	if uri, ok := v.matcher.Match(path, vars); ok {
		if uh, ok1 := v.list[uri]; ok1 {
			return uh, false
		}
	}
	if uh, ok := v.list[anyPath]; ok {
		return uh, true
	}
	return nil, false
}

// Route add new route
func (v *ctrlHandler) Route(path string, ctrl func(ctx Ctx), methods []string) {
	uh := v
	uris := urlSplit(path)
	for _, uri := range uris {
		if hasParamMatch(uri) {
			if err := uh.matcher.Add(uri); err != nil {
				panic(err)
			}
		}
		uh = uh.append(uri)
	}
	for _, m := range methods {
		uh.methods[strings.ToUpper(m)] = ctrl
	}
}

// Middlewares add middleware to route
func (v *ctrlHandler) Middlewares(path string, middlewares ...Middleware) {
	uh := v
	uris := urlSplit(path)
	for _, uri := range uris {
		uh = uh.append(uri)
	}
	uh.middlewares = append(uh.middlewares, middlewares...)
}

func (v *ctrlHandler) NoFoundHandler(call func(ctx Ctx)) {
	v.notFound = call
}

// Match find route in tree
func (v *ctrlHandler) Match(path string, method string) (int, func(ctx Ctx), uriParamData, []Middleware) {
	uris := urlSplit(path)

	fork := v
	midd := append(make([]Middleware, 0, len(fork.middlewares)), fork.middlewares...)

	vr := uriParamData{}
	var isBreak bool
	for _, uri := range uris {
		if fork, isBreak = fork.next(uri, vr); fork != nil {
			midd = append(midd, fork.middlewares...)
			if isBreak {
				break
			}
			continue
		}
		if v.notFound != nil {
			return http.StatusOK, v.notFound, nil, midd
		}
		return http.StatusNotFound, nil, nil, v.middlewares
	}
	if ctrl, ok := fork.methods[method]; ok {
		return http.StatusOK, ctrl, vr, midd
	}
	if v.notFound != nil {
		return http.StatusOK, v.notFound, nil, midd
	}
	if len(fork.methods) == 0 {
		return http.StatusNotFound, nil, nil, v.middlewares
	}
	return http.StatusMethodNotAllowed, nil, nil, v.middlewares
}
