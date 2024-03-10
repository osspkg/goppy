/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type Formatter interface {
	Encode(m *Message) ([]byte, error)
}

type FormatJSON struct{}

func NewFormatJSON() *FormatJSON {
	return &FormatJSON{}
}

func (*FormatJSON) Encode(m *Message) ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

var poolBuff = sync.Pool{
	New: func() interface{} {
		return newBuff()
	},
}

func newBuff() *bytes.Buffer {
	return bytes.NewBuffer(nil)
}

type FormatString struct {
	delim string
}

func NewFormatString() *FormatString {
	return &FormatString{delim: "\t"}
}

func (v *FormatString) SetDelimiter(d string) {
	v.delim = d
}

func (v *FormatString) Encode(m *Message) ([]byte, error) {
	b, ok := poolBuff.Get().(*bytes.Buffer)
	if !ok {
		b = newBuff()
	}

	defer func() {
		b.Reset()
		poolBuff.Put(b)
	}()

	fmt.Fprintf(b, "time=%s%slvl=%s%smsg=%#v", time.Unix(m.UnixTime, 0).Format(time.RFC3339), v.delim, m.Level, v.delim, m.Message)
	if len(m.Ctx) > 0 {
		for key, value := range m.Ctx {
			fmt.Fprintf(b, "%s%s=%#v", v.delim, key, value)
		}
	}
	b.WriteString("\n")

	return append(make([]byte, 0, b.Len()), b.Bytes()...), nil
}
