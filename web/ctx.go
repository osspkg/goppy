/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

//go:generate easyjson

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"go.osspkg.com/ioutils"
	"go.osspkg.com/ioutils/data"
	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v3/web/encoders"
)

type (
	_ctx struct {
		w http.ResponseWriter
		r *http.Request
	}

	// Ctx request and response interface
	Ctx interface {
		URL() *url.URL
		Redirect(uri string)
		Param(key string) Param
		Query(key string) string
		Header() Header
		Cookie() Cookie

		BindRaw() (*data.Buffer, error)
		BindBytes(in *[]byte) error
		BindJSON(in json.Unmarshaler) error
		BindXML(in any) error

		Bytes(code int, b []byte)
		String(code int, b string, args ...any)
		JSON(code int, in json.Marshaler)
		Stream(code int, in []byte, filename string)
		StreamFile(code int, in io.Reader, filename string)

		Error(code int, err error)
		ErrorJSON(code int, err error, args ...any)

		Context() context.Context
		SetContextValue(key, value any)
		GetContextValue(key any) any

		Request() *http.Request
		Response() http.ResponseWriter
	}
)

func NewCtx(w http.ResponseWriter, r *http.Request) Ctx {
	return &_ctx{
		w: w,
		r: r,
	}
}

/**********************************************************************************************************************/

func (v *_ctx) Request() *http.Request {
	return v.r
}

func (v *_ctx) Response() http.ResponseWriter {
	return v.w
}

/**********************************************************************************************************************/

type (
	// Param interface for typing a parameter from a URL
	Param interface {
		String() (string, error)
		Int() (int64, error)
		Float() (float64, error)
	}
	_param struct {
		val string
		err error
	}
)

// String getting the parameter as a string
func (v _param) String() (string, error) { return v.val, v.err }

// Int getting the parameter as a int64
func (v _param) Int() (int64, error) {
	if v.err != nil {
		return 0, v.err
	}
	return strconv.ParseInt(v.val, 10, 64)
}

// Float getting the parameter as a float64
func (v _param) Float() (float64, error) {
	if v.err != nil {
		return 0.0, v.err
	}
	return strconv.ParseFloat(v.val, 64)
}

// Param getting a parameter from URL by key
func (v *_ctx) Param(key string) Param {
	val, err := ParamString(v.r, key)
	return _param{
		val: val,
		err: err,
	}
}

// Query getting a query from URL by key
func (v *_ctx) Query(key string) string {
	return v.URL().Query().Get(key)
}

/**********************************************************************************************************************/

type (
	Header interface {
		Get(key string) string
		Set(key, value string)
		Del(key string)
		Val(key string) string
	}

	_header struct {
		r http.Header
		w http.Header
	}
)

func (v *_header) Get(key string) string {
	return v.r.Get(key)
}

func (v *_header) Set(key, value string) {
	v.w.Set(key, value)
}

func (v *_header) Del(key string) {
	v.w.Del(key)
}

func (v *_header) Val(key string) string {
	return v.w.Get(key)
}

func (v *_ctx) Header() Header {
	return &_header{
		r: v.r.Header,
		w: v.w.Header(),
	}
}

/**********************************************************************************************************************/

type (
	Cookie interface {
		Get(key string) string
		Set(value *http.Cookie)
	}

	_cookie struct {
		r *http.Request
		w http.ResponseWriter
	}
)

// Get getting cookies from a key request
func (v *_cookie) Get(key string) string {
	c, err := v.r.Cookie(key)
	if err != nil {
		return ""
	}
	return c.Value
}

// Set setting cookies in response
func (v *_cookie) Set(value *http.Cookie) {
	http.SetCookie(v.w, value)
}

func (v *_ctx) Cookie() Cookie {
	return &_cookie{
		r: v.r,
		w: v.w,
	}
}

/**********************************************************************************************************************/

func (v *_ctx) BindBytes(in *[]byte) error {
	b, err := ioutils.ReadAll(v.r.Body)
	if err != nil {
		return err
	}
	*in = append(*in, b...)
	return nil
}

func (v *_ctx) BindRaw() (*data.Buffer, error) {
	defer func() {
		v.r.Body.Close() //nolint:errcheck
	}()
	buf := data.NewBuffer(128)
	if _, err := buf.ReadFrom(v.r.Body); err != nil {
		return nil, err
	}
	return buf, nil
}

func (v *_ctx) BindJSON(in json.Unmarshaler) error {
	return encoders.JSONDecode(v.r, in)
}

func (v *_ctx) BindXML(in any) error {
	return encoders.XMLDecode(v.r, in)
}

func (v *_ctx) BindFormData(maxMemory int64, in any) error {
	return encoders.FormDataDecode(v.r, maxMemory, in)
}

/**********************************************************************************************************************/

//easyjson:json
type errMessage struct {
	Message string         `json:"msg"`
	Ctx     map[string]any `json:"ctx,omitempty"`
}

func (v *_ctx) ErrorJSON(code int, err error, args ...any) {
	if err == nil {
		err = fmt.Errorf("unknown error")
	}

	model := errMessage{Message: err.Error()}

	if len(args) > 0 {
		if len(args)%2 != 0 {
			args = append(args, "<unknown>")
		}
		model.Ctx = make(map[string]any, len(args)/2)
		for arg := range slices.Chunk(args, 2) {
			if len(arg) != 2 {
				continue
			}
			model.Ctx[typingJSONKey(arg[0])] = typingJSONValue(arg[1])
		}
	}

	encoders.JSONEncode(v.w, code, &model)
}

func (v *_ctx) Error(code int, err error) {
	encoders.ErrorEncode(v.w, code, err)
}

func (v *_ctx) Bytes(code int, b []byte) {
	encoders.BytesEncode(v.w, code, b)
}

// String recording the response in string format
func (v *_ctx) String(code int, b string, args ...any) {
	encoders.StringEncode(v.w, code, b, args...)
}

// JSON recording the response in json format
func (v *_ctx) JSON(code int, in json.Marshaler) {
	encoders.JSONEncode(v.w, code, in)
}

// Stream sending raw data in response with the definition of the content type by the file name
func (v *_ctx) Stream(code int, in []byte, filename string) {
	encoders.StreamEncode(v.w, code, in, filename)
}

func (v *_ctx) StreamFile(code int, in io.Reader, filename string) {
	encoders.ReaderEncode(v.w, code, in, filename)
}

/**********************************************************************************************************************/

// Context provider the request context
func (v *_ctx) Context() context.Context {
	return v.r.Context()
}

func (v *_ctx) SetContextValue(key, value any) {
	defer func() {
		if err := recover(); err != nil {
			logx.Error(
				"web.SetContextValue",
				"key", fmt.Sprintf("%#v", key),
				"err", fmt.Sprintf("%+v", err),
			)
		}
	}()
	ctx := context.WithValue(v.r.Context(), key, value)
	v.r = v.r.WithContext(ctx)
}

func (v *_ctx) GetContextValue(key any) any {
	defer func() {
		if err := recover(); err != nil {
			logx.Error(
				"web.GetContextValue",
				"key", fmt.Sprintf("%#v", key),
				"err", fmt.Sprintf("%+v", err),
			)
		}
	}()

	return v.r.Context().Value(key)
}

/**********************************************************************************************************************/

// URL getting a URL from a request
func (v *_ctx) URL() *url.URL {
	uri := v.r.URL
	uri.Host = v.r.Host
	return uri
}

/**********************************************************************************************************************/

// Redirect redirecting to another URL
func (v *_ctx) Redirect(uri string) {
	http.Redirect(v.w, v.r, uri, http.StatusMovedPermanently)
}
