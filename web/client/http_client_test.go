/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/errors"

	"go.osspkg.com/goppy/v2/auth/signature"
	"go.osspkg.com/goppy/v2/web/client"
	"go.osspkg.com/goppy/v2/web/encoders"
)

type mockModelName struct {
	Data struct {
		Name string `json:"name"`
	}
}

func (v *mockModelName) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &v.Data)
}

func (v *mockModelName) MarshalJSON() ([]byte, error) {
	return json.Marshal(&v.Data)
}

type mockHandler struct {
}

func (*mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	b, err := httputil.DumpRequest(r, true)
	fmt.Println("----------------------------------------------------")
	fmt.Println(string(b))
	fmt.Println("----------------------------------------------------")
	fmt.Println("ERR DUMP:", err)

	switch r.URL.Path {
	case "/empty":
		w.WriteHeader(http.StatusOK)

	case "/out-only":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"test"}`))

	case "/form-data":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestUnit_HTTPClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	srv := &http.Server{
		Addr:           "127.0.0.1:12345",
		Handler:        &mockHandler{},
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			casecheck.NoError(t, err)
		}
	}()
	defer func() {
		casecheck.NoError(t, srv.Close())
	}()
	defer cancel()

	time.Sleep(time.Second * 2)

	cli := client.NewHTTPClient(
		client.WithDefaultHeaders(map[string]string{
			"User-Agent": "Mozilla/5.0",
		}),
		client.WithSignatures(map[string]signature.Signature{
			"127.0.0.1:12345": signature.NewSHA1("key-1", "1234567890"),
		}),
	)

	err := cli.Send(ctx, http.MethodPost, "http://127.0.0.1:12345/empty", nil, nil)
	casecheck.NoError(t, err)

	var actual mockModelName
	err = cli.Send(ctx, http.MethodPost, "http://127.0.0.1:12345/out-only", &mockModelName{}, &actual)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "test", actual.Data.Name)

	err = cli.Send(ctx, http.MethodPost, "http://127.0.0.1:12345/fail", &mockModelName{}, nil)
	casecheck.Error(t, err)
	casecheck.ErrorContains(t, err, "http client: bad status code: 500")
	he, ok := err.(*client.HTTPError)
	casecheck.True(t, ok)
	casecheck.Equal(t, http.StatusInternalServerError, he.Code)
	casecheck.Equal(t, "", he.ContentType)
	casecheck.NotNil(t, he.Raw)
	casecheck.Equal(t, "", he.Raw.String())

	fd := &encoders.FormData{}
	fd.Field("text", "test")
	fd.Field("number", 1234)
	fd.Field("json", &mockModelName{})
	err = cli.Send(ctx, http.MethodPost, "http://127.0.0.1:12345/form-data", fd, nil)
	casecheck.NoError(t, err)
}
