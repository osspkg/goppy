/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"errors"
	"fmt"
)

//go:generate easyjson

var (
	ErrParseJSON      = &Error{Code: -32700, Message: "json parse error"}
	ErrInvalidRequest = &Error{Code: -32600, Message: "invalid request"}
	ErrMethodNotFound = &Error{Code: -32601, Message: "method not found"}
	ErrInvalidParams  = &Error{Code: -32602, Message: "invalid params"}
	ErrInternal       = &Error{Code: -32603, Message: "internal error"}
)

//easyjson:json
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []any  `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func matchError(e error) *Error {
	var err *Error
	if errors.As(e, &err) {
		return err
	}
	return &Error{
		Message: e.Error(),
	}
}
