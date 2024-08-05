/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package event

//go:generate easyjson

import (
	"encoding/json"
	"fmt"

	"go.osspkg.com/ioutils/pool"
)

var (
	poolEvents = pool.New[*event](func() *event { return &event{} })
)

type Id uint16

type (
	Event interface {
		ID() Id
		Decode(out interface{}) error
		Encode(in interface{}) error
		Reset()
		WithError(e error)
		WithID(id Id)
	}

	//easyjson:json
	event struct {
		Id   Id              `json:"e"`
		Data json.RawMessage `json:"d,omitempty"`
		Err  *string         `json:"err,omitempty"`
	}
)

func New(call func(ev Event)) {
	m := poolEvents.Get()
	call(m)
	poolEvents.Put(m)
}

func (v *event) ID() Id {
	return v.Id
}

func (v *event) Decode(in interface{}) error {
	if v.Err != nil {
		return fmt.Errorf("%s", *v.Err)
	}
	return json.Unmarshal(v.Data, in)
}

func (v *event) Encode(in interface{}) (err error) {
	v.Data, err = json.Marshal(in)
	if err != nil {
		return err
	}
	v.Err = nil
	return nil
}

func (v *event) WithID(id Id) {
	v.Id = id
}

func (v *event) Reset() {
	v.Id, v.Err, v.Data = 0, nil, v.Data[:0]
}

func (v *event) WithError(err error) {
	if err == nil {
		return
	}
	errMsg := err.Error()
	v.Err, v.Data = &errMsg, v.Data[:0]
}
