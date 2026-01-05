/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/ioutils/data"

	"go.osspkg.com/goppy/v3/web/encoders"
)

type B struct {
	Name string `json:"name"`
}

type jsonModel struct {
	B B
}

func (m *jsonModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.B)
}

func (m *jsonModel) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &m.B)
}

func TestUnit_FormData(t *testing.T) {
	formData := &encoders.FormData{}
	formData.Field("name", "UserName")
	formData.Field("count", 123)
	formData.Field("bool", true)
	formData.Field("json_model", &jsonModel{B: B{Name: "dd"}})
	formData.File("file", "hello.txt", bytes.NewBufferString("hello world"))

	casecheck.Nil(t, formData.Reader())
	casecheck.Equal(t, "", formData.ContentType())

	casecheck.NoError(t, formData.Encode())

	req := httptest.NewRequest(http.MethodPost, "/", formData.Reader())
	req.Header.Set("Content-Type", formData.ContentType())

	casecheck.Error(t, encoders.FormDataDecode(req, 1024*100, nil))

	type A struct {
		Name     string        `formData:"name"`
		NamePtr  *string       `formData:"name"`
		Count    int           `formData:"count"`
		CountPtr *int          `formData:"count,omitempty"`
		CountNil *int          `formData:"count_nil,omitempty"`
		Bool     bool          `formData:"bool,omitempty"`
		BoolPtr  *bool         `formData:"bool,omitempty"`
		FileIOR  io.Reader     `formData:"file"`
		FileIORS io.ReadSeeker `formData:"file,omitempty"`
		Json     jsonModel     `formData:"json_model,omitempty"`
		JsonPtr  *jsonModel    `formData:"json_model,omitempty"`
		Buff     data.Buffer   `formData:"file,omitempty"`
		BuffPtr  *data.Buffer  `formData:"file,omitempty"`
	}

	buf := data.NewBuffer(1024)

	a := A{
		FileIOR: buf,
		BuffPtr: buf,
		JsonPtr: &jsonModel{},
	}
	casecheck.NoError(t, encoders.FormDataDecode(req, 1024*100, &a))

	casecheck.Equal(t, "UserName", a.Name)
	casecheck.Equal(t, "UserName", *a.NamePtr)

	casecheck.Equal(t, 123, a.Count)
	casecheck.Equal(t, 123, *a.CountPtr)
	casecheck.Nil(t, a.CountNil)

	casecheck.True(t, a.Bool)
	casecheck.True(t, *a.BoolPtr)

	casecheck.NotNil(t, a.FileIOR)
	casecheck.NotNil(t, a.FileIORS)
	b, err := io.ReadAll(a.FileIOR)
	casecheck.NoError(t, err)
	casecheck.Equal(t, "hello worldhello world", string(b))

	casecheck.Equal(t, "dd", a.Json.B.Name)
	casecheck.Equal(t, "dd", a.JsonPtr.B.Name)
	casecheck.Equal(t, "hello world", a.Buff.String())
	casecheck.Equal(t, "hello worldhello world", buf.String())
}

func Benchmark_UnmarshalForm(b *testing.B) {
	formData := &encoders.FormData{}
	formData.Field("name", "UserName")
	formData.Field("count", 123)
	formData.Field("bool", true)
	formData.Field("json_model", &jsonModel{B: B{Name: "dd"}})
	formData.File("file", "hello.txt", bytes.NewBufferString("hello world"))
	casecheck.NoError(b, formData.Encode())

	req := httptest.NewRequest(http.MethodPost, "/", formData.Reader())
	req.Header.Set("Content-Type", formData.ContentType())

	type A struct {
		Name     string        `formData:"name"`
		NamePtr  *string       `formData:"name"`
		Count    int           `formData:"count"`
		CountPtr *int          `formData:"count,omitempty"`
		CountNil *int          `formData:"count_nil,omitempty"`
		Bool     bool          `formData:"bool,omitempty"`
		BoolPtr  *bool         `formData:"bool,omitempty"`
		FileIOR  io.Reader     `formData:"file"`
		FileIORS io.ReadSeeker `formData:"file,omitempty"`
		Json     jsonModel     `formData:"json_model,omitempty"`
		JsonPtr  *jsonModel    `formData:"json_model,omitempty"`
		Buff     data.Buffer   `formData:"file,omitempty"`
		BuffPtr  *data.Buffer  `formData:"file,omitempty"`
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var a A
		if e := encoders.FormDataDecode(req, 1024*100, &a); e != nil {
			b.Fatal(e)
		}
	}
}
