/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

//go:generate easyjson
import (
	"encoding/json"
)

//easyjson:json
type request struct {
	Id     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

//easyjson:json
type bulkRequest []request

func (br *bulkRequest) Reset() {
	*br = (*br)[:0]
}

//easyjson:json
type response struct {
	Id     string       `json:"id"`
	Result any          `json:"result,omitempty"`
	Error  *errResponse `json:"error,omitempty"`
}

//easyjson:json
type bulkResponse []response

//easyjson:json
type errResponse struct {
	Message string            `json:"message"`
	Code    int64             `json:"code"`
	Ctx     map[string]string `json:"ctx,omitempty"`
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
