/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"
	"testing"

	"go.osspkg.com/goppy/xtest"
)

func TestUnit_NewHandler(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/aaa/{id}", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})
	h.Route("", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})

	code, ctrl, vr, midd := h.Match("/aaa/bbb", http.MethodPost)
	xtest.Equal(t, 200, code)
	xtest.NotNil(t, ctrl)
	xtest.Equal(t, 0, len(midd))
	xtest.Equal(t, uriParamData{"id": "bbb"}, vr)

	h.Middlewares("/aaa", RecoveryMiddleware(nil))
	h.Middlewares("", RecoveryMiddleware(nil))

	code, ctrl, vr, midd = h.Match("/aaa/ccc", http.MethodGet)
	xtest.Equal(t, http.StatusMethodNotAllowed, code)
	xtest.Nil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/aaa/bbb", http.MethodPost)
	xtest.Equal(t, http.StatusOK, code)
	xtest.NotNil(t, ctrl)
	xtest.Equal(t, 2, len(midd))
	xtest.Equal(t, uriParamData{"id": "bbb"}, vr)

	code, ctrl, vr, midd = h.Match("", http.MethodPost)
	xtest.Equal(t, http.StatusOK, code)
	xtest.NotNil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData{}, vr)

	h.Middlewares("/www/www/www", RecoveryMiddleware(nil))

	code, ctrl, vr, midd = h.Match("/www/www/www", http.MethodPost)
	xtest.Equal(t, http.StatusNotFound, code)
	xtest.Nil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	xtest.Equal(t, http.StatusNotFound, code)
	xtest.Nil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData(nil), vr)

	h.NoFoundHandler(func(_ http.ResponseWriter, _ *http.Request) {})

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	xtest.Equal(t, http.StatusOK, code)
	xtest.NotNil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData(nil), vr)
}

func TestUnit_NewHandler2(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/api/v{id}/data/#", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodGet})

	h.Middlewares("/api/v{id}", RecoveryMiddleware(nil))

	code, ctrl, vr, midd := h.Match("/api/v1/data/user/aaaa", http.MethodGet)
	xtest.Equal(t, http.StatusOK, code)
	xtest.NotNil(t, ctrl)
	xtest.Equal(t, 1, len(midd))
	xtest.Equal(t, uriParamData{"id": "1"}, vr)

}
