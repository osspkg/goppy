/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"encoding/json"
)

//go:generate easyjson

//easyjson:json
type RequestBatch []Request

//easyjson:json
type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

//easyjson:json
type ResponseBatch []Response

//easyjson:json
type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
}

type Batch struct {
	Method   string
	Params   Jsoner
	Fallback func(json.RawMessage, *Error)
}
