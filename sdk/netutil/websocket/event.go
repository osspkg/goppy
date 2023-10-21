/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

//go:generate easyjson

import (
	"encoding/json"
	"sync"
)

var (
	poolEvents = sync.Pool{New: func() interface{} { return &event{} }}
)

type EventID uint16

//easyjson:json
type event struct {
	ID   EventID         `json:"e"`
	Data json.RawMessage `json:"d"`
	Err  *string         `json:"err,omitempty"`
}

func getEventModel(call func(ev *event)) {
	m, ok := poolEvents.Get().(*event)
	if !ok {
		m = &event{}
	}
	call(m)
	poolEvents.Put(m.Reset())
}

func (v *event) EventID() EventID {
	return v.ID
}

func (v *event) Decode(in interface{}) error {
	return json.Unmarshal(v.Data, in)
}

func (v *event) Encode(in interface{}) {
	b, err := json.Marshal(in)
	if err != nil {
		v.Error(err)
		return
	}
	v.Body(b)
}

func (v *event) EncodeEvent(id EventID, in interface{}) {
	v.ID = id
	v.Encode(in)
}

func (v *event) Reset() *event {
	v.ID, v.Err, v.Data = 0, nil, v.Data[:0]
	return v
}

func (v *event) Error(e error) {
	if e == nil {
		return
	}
	err := e.Error()
	v.Err, v.Data = &err, v.Data[:0]
}

func (v *event) Body(b []byte) {
	v.Err, v.Data = nil, append(v.Data[:0], b...)
}
