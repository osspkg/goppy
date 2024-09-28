/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/goppy/v2/web"
)

func TestUnit_Route1(t *testing.T) {
	result := new(string)
	r := web.NewBaseRouter()
	r.Global(func(c func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			*result += "1"
			c(w, r)
		}
	})
	r.Global(func(c func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			*result += "2"
			c(w, r)
		}
	})
	r.Global(func(c func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			*result += "3"
			c(w, r)
		}
	})
	r.Route("/", func(w http.ResponseWriter, r *http.Request) {
		*result += "Ctrl"
	}, http.MethodGet)
	r.Middlewares("/test", func(c func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			*result += "4"
			c(w, r)
		}
	})
	r.Middlewares("/", func(c func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			*result += "5"
			c(w, r)
		}
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	casecheck.Equal(t, "1235Ctrl", *result)
}

type statusInterface interface {
	Result() *http.Response
}

func getStatusAndClose(s statusInterface) int {
	resp := s.Result()
	code := resp.StatusCode
	err := resp.Body.Close()
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	return code
}

func TestUnit_Route2(t *testing.T) {
	r := web.NewBaseRouter()
	r.Route("/{id}", func(w http.ResponseWriter, r *http.Request) {}, http.MethodGet)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/aaa/bbb/ccc/eee/ggg/fff/kkk", nil)
	r.ServeHTTP(w, req)
	casecheck.Equal(t, 404, getStatusAndClose(w))

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/aaa/", nil)
	r.ServeHTTP(w, req)
	casecheck.Equal(t, 200, getStatusAndClose(w))

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/aaa", nil)
	r.ServeHTTP(w, req)
	casecheck.Equal(t, 200, getStatusAndClose(w))

	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/aaa?a=1", nil)
	r.ServeHTTP(w, req)
	casecheck.Equal(t, 200, getStatusAndClose(w))
}

func mockNilHandler(_ http.ResponseWriter, _ *http.Request) {}

func BenchmarkRouter0(b *testing.B) {
	serv := web.NewBaseRouter()
	serv.Route(`/aaa/bbb/ccc/eee/ggg/fff/kkk`, mockNilHandler, http.MethodGet)
	serv.Route(`/aaa/bbb/000/eee/ggg/fff/kkk`, mockNilHandler, http.MethodGet)

	req := []*http.Request{
		httptest.NewRequest("GET", "/aaa/bbb/ccc/eee/ggg/fff/kkk", nil),
		httptest.NewRequest("GET", "/aaa/bbb/000/eee/ggg/fff/kkk", nil),
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			b.Run("", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					serv.ServeHTTP(w, req[i%2])
					code := getStatusAndClose(w)
					if code != http.StatusOK {
						b.Fatalf("invalid code: %d", code)
					}
					w.Flush()
				}
			})
		}
	})
}

func BenchmarkRouter1(b *testing.B) {
	serv := web.NewBaseRouter()
	serv.Route(`/{id0}/{id1}/{id2:\d+}/{id3}/{id4}/{id5}/{id6}`, mockNilHandler, http.MethodGet)
	serv.Route(`/{id0}/{id1}/{id2:\w+}/{id3}/{id4}/{id5}/{id6}`, mockNilHandler, http.MethodGet)

	req := []*http.Request{
		httptest.NewRequest("GET", "/aaa/bbb/ccc/eee/ggg/fff/kkk", nil),
		httptest.NewRequest("GET", "/aaa/bbb/000/eee/ggg/fff/kkk", nil),
	}

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			b.Run("", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					serv.ServeHTTP(w, req[i%2])
					code := getStatusAndClose(w)
					if code != http.StatusOK {
						b.Fatalf("invalid code: %d", code)
					}
					w.Flush()
				}
			})
		}
	})
}
