/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package event

//go:generate easyjson

import (
	"encoding/json"
	"sync"
)

var (
	poolEvents = sync.Pool{New: func() interface{} { return &Message{} }}
)

type Id uint16

//easyjson:json
type Message struct {
	ID   Id              `json:"e"`
	Data json.RawMessage `json:"d"`
	Err  *string         `json:"err,omitempty"`
}

func GetMessage(call func(ev *Message)) {
	m, ok := poolEvents.Get().(*Message)
	if !ok {
		m = &Message{}
	}
	call(m)
	poolEvents.Put(m.Reset())
}

func (v *Message) EventID() Id {
	return v.ID
}

func (v *Message) Decode(in interface{}) error {
	return json.Unmarshal(v.Data, in)
}

func (v *Message) Encode(in interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		v.Error(err)
		return
	}
	v.Body(b)
}

func (v *Message) EncodeEvent(id Id, in interface{}) {
	v.ID = id
	v.Encode(in)
}

func (v *Message) Reset() *Message {
	v.ID, v.Err, v.Data = 0, nil, v.Data[:0]
	return v
}

func (v *Message) Error(e error) {
	if e == nil {
		return
	}
	err := e.Error()
	v.Err, v.Data = &err, v.Data[:0]
}

func (v *Message) Body(b []byte) {
	v.Err, v.Data = nil, append(v.Data[:0], b...)
}
