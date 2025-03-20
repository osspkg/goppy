/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import "fmt"

//go:generate easyjson

//easyjson:json
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func createErrorResponseBatch(err error, reqs RequestBatch) ResponseBatch {
	resp := make(ResponseBatch, 0, len(reqs))
	for _, req := range reqs {
		resp = append(resp, newErrorResponse(err, req))
	}
	return resp
}

func newErrorResponse(err error, req Request) Response {
	return Response{
		ID: req.ID, Result: nil, Error: &Error{
			Message: err.Error(),
		},
	}
}
