/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_NewHandler(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/aaa/{id}", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})
	h.Route("", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})

	code, ctrl, vr, midd := h.Match("/aaa/bbb", http.MethodPost)
	casecheck.Equal(t, 200, code)
	casecheck.NotNil(t, ctrl)
	casecheck.Equal(t, 0, len(midd))
	casecheck.Equal(t, uriParamData{"id": "bbb"}, vr)

	h.Middlewares("/aaa", RecoveryMiddleware())
	h.Middlewares("", RecoveryMiddleware())

	code, ctrl, vr, midd = h.Match("/aaa/ccc", http.MethodGet)
	casecheck.Equal(t, http.StatusMethodNotAllowed, code)
	casecheck.Nil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/aaa/bbb", http.MethodPost)
	casecheck.Equal(t, http.StatusOK, code)
	casecheck.NotNil(t, ctrl)
	casecheck.Equal(t, 2, len(midd))
	casecheck.Equal(t, uriParamData{"id": "bbb"}, vr)

	code, ctrl, vr, midd = h.Match("", http.MethodPost)
	casecheck.Equal(t, http.StatusOK, code)
	casecheck.NotNil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData{}, vr)

	h.Middlewares("/www/www/www", RecoveryMiddleware())

	code, ctrl, vr, midd = h.Match("/www/www/www", http.MethodPost)
	casecheck.Equal(t, http.StatusNotFound, code)
	casecheck.Nil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	casecheck.Equal(t, http.StatusNotFound, code)
	casecheck.Nil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData(nil), vr)

	h.NoFoundHandler(func(_ http.ResponseWriter, _ *http.Request) {})

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	casecheck.Equal(t, http.StatusOK, code)
	casecheck.NotNil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData(nil), vr)
}

func TestUnit_NewHandler2(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/api/v{id}/data/#", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodGet})

	h.Middlewares("/api/v{id}", RecoveryMiddleware())

	code, ctrl, vr, midd := h.Match("/api/v1/data/user/aaaa", http.MethodGet)
	casecheck.Equal(t, http.StatusOK, code)
	casecheck.NotNil(t, ctrl)
	casecheck.Equal(t, 1, len(midd))
	casecheck.Equal(t, uriParamData{"id": "1"}, vr)

}
