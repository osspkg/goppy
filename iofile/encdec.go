/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iofile

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
	"gopkg.in/yaml.v3"
)

var (
	errBadFileFormat = errors.New("format is not a supported")

	fileCodec = newCodec().
			Add(".yml", yaml.Marshal, yaml.Unmarshal).
			Add(".yaml", yaml.Marshal, yaml.Unmarshal).
			Add(".json", json.Marshal, json.Unmarshal)
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type codec struct {
	enc map[string]func(v interface{}) ([]byte, error)
	dec map[string]func([]byte, interface{}) error
	mux iosync.Lock
}

func newCodec() *codec {
	return &codec{
		enc: make(map[string]func(v interface{}) ([]byte, error), 10),
		dec: make(map[string]func([]byte, interface{}) error, 10),
		mux: iosync.NewLock(),
	}
}

func AddFileCodec(ext string, enc func(v interface{}) ([]byte, error), dec func([]byte, interface{}) error) {
	fileCodec.Add(ext, enc, dec)
}

func (v *codec) Add(ext string, enc func(v interface{}) ([]byte, error), dec func([]byte, interface{}) error) *codec {
	v.mux.Lock(func() {
		v.enc[ext] = enc
		v.dec[ext] = dec
	})
	return v
}

func (v *codec) GetEnc(ext string) (fn func(v interface{}) ([]byte, error), ok bool) {
	v.mux.RLock(func() {
		fn, ok = v.enc[ext]
	})
	return
}

func (v *codec) GetDec(ext string) (fn func([]byte, interface{}) error, ok bool) {
	v.mux.RLock(func() {
		fn, ok = v.dec[ext]
	})
	return
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type FileCodec string

func (v FileCodec) Decode(configs ...interface{}) error {
	data, err := os.ReadFile(string(v))
	if err != nil {
		return err
	}
	ext := filepath.Ext(string(v))
	c, ok := fileCodec.GetDec(ext)
	if !ok {
		return errBadFileFormat
	}
	return v.dec(data, c, configs...)
}

func (v FileCodec) Encode(configs ...interface{}) error {
	ext := filepath.Ext(string(v))
	c, ok := fileCodec.GetEnc(ext)
	if !ok {
		return errBadFileFormat
	}
	b, err := v.enc(c, configs...)
	if err != nil {
		return err
	}
	return os.WriteFile(string(v), b, 0755)
}

func (v FileCodec) dec(data []byte, call func([]byte, interface{}) error, configs ...interface{}) error {
	for _, conf := range configs {
		if err := call(data, conf); err != nil {
			return err
		}
	}
	return nil
}

func (v FileCodec) enc(call func(v interface{}) ([]byte, error), configs ...interface{}) ([]byte, error) {
	b := make([]byte, 0, 300*len(configs))
	for _, conf := range configs {
		bb, err := call(conf)
		if err != nil {
			return nil, err
		}
		b = append(b, '\n', '\n')
		b = append(b, bb...)
	}
	return b, nil
}
