/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/web/encoders"
)

func TestUnit_FormData(t *testing.T) {
	formData := &encoders.FormData{}
	formData.Field("name", "UserName")
	formData.Field("count", 123)
	formData.Field("bool", true)
	formData.File("file", "hello.txt", bytes.NewBufferString("hello world"))

	casecheck.Nil(t, formData.Reader())
	casecheck.Equal(t, "", formData.ContentType())

	casecheck.NoError(t, formData.Encode())

	req := httptest.NewRequest(http.MethodPost, "/", formData.Reader())
	req.Header.Set("Content-Type", formData.ContentType())

	casecheck.Error(t, encoders.FormDataDecode(req, 1024*100, nil))

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
	casecheck.NoError(t, encoders.FormDataDecode(req, 1024*100, &a))

	casecheck.Equal(t, "UserName", a.Name)
	casecheck.Equal(t, "UserName", *a.NameNil)
	casecheck.Equal(t, 123, a.Count)
	casecheck.Nil(t, a.CountNil)
	casecheck.True(t, a.Bool)
	casecheck.Nil(t, a.BoolNil)
	casecheck.NotNil(t, a.FileRSC)
	casecheck.Nil(t, a.FileR)

	b, err := io.ReadAll(a.FileRSC)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "hello world", string(b))
}
