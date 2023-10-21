/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package webutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnit_NewHandler(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/aaa/{id}", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})
	h.Route("", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodPost})

	code, ctrl, vr, midd := h.Match("/aaa/bbb", http.MethodPost)
	require.Equal(t, 200, code)
	require.NotNil(t, ctrl)
	require.Equal(t, 0, len(midd))
	require.Equal(t, uriParamData{"id": "bbb"}, vr)

	h.Middlewares("/aaa", RecoveryMiddleware(nil))
	h.Middlewares("", RecoveryMiddleware(nil))

	code, ctrl, vr, midd = h.Match("/aaa/ccc", http.MethodGet)
	require.Equal(t, http.StatusMethodNotAllowed, code)
	require.Nil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/aaa/bbb", http.MethodPost)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, ctrl)
	require.Equal(t, 2, len(midd))
	require.Equal(t, uriParamData{"id": "bbb"}, vr)

	code, ctrl, vr, midd = h.Match("", http.MethodPost)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData{}, vr)

	h.Middlewares("/www/www/www", RecoveryMiddleware(nil))

	code, ctrl, vr, midd = h.Match("/www/www/www", http.MethodPost)
	require.Equal(t, http.StatusNotFound, code)
	require.Nil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData(nil), vr)

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	require.Equal(t, http.StatusNotFound, code)
	require.Nil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData(nil), vr)

	h.NoFoundHandler(func(_ http.ResponseWriter, _ *http.Request) {})

	code, ctrl, vr, midd = h.Match("/test", http.MethodGet)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData(nil), vr)
}

func TestUnit_NewHandler2(t *testing.T) {
	h := newCtrlHandler()
	h.Route("/api/v{id}/data/#", func(_ http.ResponseWriter, _ *http.Request) {}, []string{http.MethodGet})

	h.Middlewares("/api/v{id}", RecoveryMiddleware(nil))

	code, ctrl, vr, midd := h.Match("/api/v1/data/user/aaaa", http.MethodGet)
	require.Equal(t, http.StatusOK, code)
	require.NotNil(t, ctrl)
	require.Equal(t, 1, len(midd))
	require.Equal(t, uriParamData{"id": "1"}, vr)

}
