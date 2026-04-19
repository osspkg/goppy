/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

//go:generate easyjson

import (
	"encoding/json"
	"fmt"
)

//easyjson:json
type requestRaw struct {
	Id     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

//easyjson:json
type bulkRequestRaw []requestRaw

func (br *bulkRequestRaw) Reset() {
	*br = (*br)[:0]
}

//easyjson:json
type requestAny struct {
	Id     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

//easyjson:json
type bulkRequestAny []requestAny

func (br *bulkRequestAny) Reset() {
	*br = (*br)[:0]
}

//easyjson:json
type responseRaw struct {
	Id     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *errResponse    `json:"error,omitempty"`
}

//easyjson:json
type bulkResponseRaw []responseRaw

func (br *bulkResponseRaw) Reset() {
	*br = (*br)[:0]
}

//easyjson:json
type responseAny struct {
	Id     string       `json:"id"`
	Result any          `json:"result,omitempty"`
	Error  *errResponse `json:"error,omitempty"`
}

//easyjson:json
type bulkResponseAny []responseAny

func (br *bulkResponseAny) Reset() {
	*br = (*br)[:0]
}

//easyjson:json
type errResponse struct {
	Message string            `json:"message"`
	Code    int64             `json:"code"`
	Ctx     map[string]string `json:"ctx,omitempty"`
}

func (e *errResponse) Error() string {
	if e.Ctx != nil {
		return fmt.Sprintf("#%d %v {%+v}", e.Code, e.Message, e.Ctx)
	}
	return fmt.Sprintf("#%d %v", e.Code, e.Message)
}

func errorConvert(e error) *errResponse {
	if e == nil {
		return nil
	}
	err := &errResponse{}
	te, ok := e.(TError)
	if ok {
		err.Code = te.GetCode()
		err.Message = te.GetMessage()
		err.Ctx = te.GetContext()
	} else {
		err.Message = e.Error()
	}
	return err
}
