/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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
	Pool = pool.New[*entity](func() *entity { return &entity{} })
)

type Id uint16

type (
	Event interface {
		ID() Id
		Decode(out any) error
		Encode(in any) error
		Reset()
		WithError(e error)
		WithID(id Id)
	}

	//easyjson:json
	entity struct {
		Id   Id              `json:"e"`
		Data json.RawMessage `json:"d,omitempty"`
		Err  *string         `json:"err,omitempty"`
	}
)

func (v *entity) ID() Id {
	return v.Id
}

func (v *entity) Decode(in any) error {
	if v.Err != nil {
		return fmt.Errorf("%s", *v.Err)
	}
	return json.Unmarshal(v.Data, in)
}

func (v *entity) Encode(in any) (err error) {
	v.Data, err = json.Marshal(in)
	if err != nil {
		return err
	}
	v.Err = nil
	return nil
}

func (v *entity) WithID(id Id) {
	v.Id = id
}

func (v *entity) Reset() {
	v.Id, v.Err, v.Data = 0, nil, v.Data[:0]
}

func (v *entity) WithError(err error) {
	if err == nil {
		return
	}
	errMsg := err.Error()
	v.Err, v.Data = &errMsg, v.Data[:0]
}
