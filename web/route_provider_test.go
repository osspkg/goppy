/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/ioutils"
)

func Test_newRouter(t *testing.T) {
	r := newRouter("test", Config{})

	r.NotFoundHandler(func(ctx Ctx) {
		ctx.String(404, "NotFoundHandler")
	})

	// nolint: lll
	api1 := r.Collection("api/", func(f func(Ctx)) func(Ctx) {
		return func(ctx Ctx) {
			f(ctx)
			fmt.Fprintf(ctx.Response(), " +(r.Collection [api/] middlewares) ")
		}
	})

	api1.Get("aaa", func(ctx Ctx) {
		ctx.String(200, "api1.Get [aaa] handler")
	})

	// nolint: lll
	api2 := api1.Collection("/bbb/ccc", func(f func(Ctx)) func(Ctx) {
		return func(ctx Ctx) {
			f(ctx)
			fmt.Fprintf(ctx.Response(), " +(api1.Collection [/bbb/ccc] middlewares) ")
		}
	})

	api2.Post("eee", func(ctx Ctx) {
		ctx.String(200, "api2.Post [aaa] handler")
	})

	buff := &bytes.Buffer{}
	requestTest(buff, r.route, http.MethodGet, "/", nil)
	requestTest(buff, r.route, http.MethodGet, "/api", nil)
	requestTest(buff, r.route, http.MethodGet, "/api/aaa", nil)
	requestTest(buff, r.route, http.MethodGet, "/api/bbb", nil)
	requestTest(buff, r.route, http.MethodGet, "/api/bbb/ccc", nil)
	requestTest(buff, r.route, http.MethodGet, "/api/bbb/ccc/eee", nil)
	requestTest(buff, r.route, http.MethodPost, "/api/bbb/ccc/eee", nil)

	expected := `GET: /
STATUS: 404
BODY: NotFoundHandler

GET: /api
STATUS: 404
BODY: NotFoundHandler +(r.Collection [api/] middlewares) 

GET: /api/aaa
STATUS: 200
BODY: api1.Get [aaa] handler +(r.Collection [api/] middlewares) 

GET: /api/bbb
STATUS: 404
BODY: NotFoundHandler +(r.Collection [api/] middlewares) 

GET: /api/bbb/ccc
STATUS: 404
BODY: NotFoundHandler +(api1.Collection [/bbb/ccc] middlewares)  +(r.Collection [api/] middlewares) 

GET: /api/bbb/ccc/eee
STATUS: 404
BODY: NotFoundHandler +(api1.Collection [/bbb/ccc] middlewares)  +(r.Collection [api/] middlewares) 

POST: /api/bbb/ccc/eee
STATUS: 200
BODY: api2.Post [aaa] handler +(api1.Collection [/bbb/ccc] middlewares)  +(r.Collection [api/] middlewares) 

`
	casecheck.Equal(t, expected, buff.String())
}

// nolint: unparam
func requestTest(buff io.Writer, handler http.Handler, method string, uri string, body io.Reader) {
	r := httptest.NewRequest(method, uri, body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	fmt.Fprintf(buff, "%s: %s\n", method, uri)
	rr := w.Result()
	defer rr.Body.Close()
	fmt.Fprintf(buff, "STATUS: %d\n", rr.StatusCode)
	b, err := ioutils.ReadAll(rr.Body)
	if err != nil {
		fmt.Fprintf(buff, "ERR: %s\n", err.Error())
		return
	}
	fmt.Fprintf(buff, "BODY: ")
	buff.Write(append(b, '\n', '\n'))
}
