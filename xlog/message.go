/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog

import "sync"

//go:generate easyjson

var poolMessage = sync.Pool{
	New: func() interface{} {
		return newMessage()
	},
}

//easyjson:json
type Message struct {
	Time    int64                  `json:"time" yaml:"time"`
	Level   string                 `json:"lvl" yaml:"lvl"`
	Message string                 `json:"msg" yaml:"msg"`
	Ctx     map[string]interface{} `json:"ctx,omitempty" yaml:"ctx,omitempty,inline"`
}

func newMessage() *Message {
	return &Message{
		Ctx: make(map[string]interface{}, 2),
	}
}

func (v *Message) Reset() {
	v.Time = 0
	v.Level = ""
	v.Message = ""
	for s := range v.Ctx {
		delete(v.Ctx, s)
	}
}
