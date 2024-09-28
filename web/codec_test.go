/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_FormDataDecode(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	mpw := multipart.NewWriter(buf)

	w, _ := mpw.CreateFormField("name")
	w.Write([]byte("UserName"))
	w, _ = mpw.CreateFormField("count")
	w.Write([]byte("123"))
	w, _ = mpw.CreateFormField("bool")
	w.Write([]byte("true"))
	w, _ = mpw.CreateFormFile("file", "hello.txt")
	w.Write([]byte("hello world"))

	casecheck.NoError(t, mpw.Close())

	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set("Content-Type", mpw.FormDataContentType())

	err := FormDataDecode(req, nil)
	casecheck.Error(t, err)

	type A struct {
		Name     string    `formData:"name"`
		NameNil  *string   `formData:"name"`
		Count    int       `formData:"count"`
		CountNil *int      `formData:"count_nil,omitempty"`
		Bool     bool      `formData:"bool,omitempty"`
		BoolNil  *bool     `formData:"bool_nil,omitempty"`
		FileRSC  io.Reader `formData:"file"`
		FileR    io.Reader `formData:"file1,omitempty"`
	}

	var a A
	err = FormDataDecode(req, &a)
	casecheck.NoError(t, err)

	casecheck.Equal(t, "UserName", a.Name)
	casecheck.Equal(t, "UserName", *a.NameNil)
	casecheck.Equal(t, 123, a.Count)
	casecheck.Nil(t, a.CountNil)
	casecheck.True(t, a.Bool)
	casecheck.Nil(t, a.BoolNil)
	casecheck.NotNil(t, a.FileRSC)
	casecheck.Nil(t, a.FileR)

	var buff bytes.Buffer
	_, err = io.Copy(&buff, a.FileRSC)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "hello world", buff.String())
}
