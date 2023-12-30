/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"bufio"
	"net"
	"net/textproto"
	"time"

	"go.osspkg.com/goppy/errors"
)

type connProcessor struct {
	conn    net.Conn
	pipe    *textproto.Reader
	buff    *bufio.Reader
	timeout time.Duration
}

func newConnProcessor(c net.Conn, ttl time.Duration) *connProcessor {
	buff := bufio.NewReader(c)
	return &connProcessor{
		pipe:    textproto.NewReader(buff),
		buff:    buff,
		conn:    c,
		timeout: ttl,
	}
}

func (v *connProcessor) ReadLine() ([]byte, error) {
	if err := v.updateDeadline(); err != nil {
		return nil, err
	}
	return v.pipe.ReadLineBytes()
}

func (v *connProcessor) Read(b []byte) (int, error) {
	if err := v.updateDeadline(); err != nil {
		return 0, err
	}
	return v.buff.Read(b)
}

func (v *connProcessor) Addr() net.Addr {
	return v.conn.RemoteAddr()
}

func (v *connProcessor) Write(b []byte) (int, error) {
	return v.conn.Write(b)
}

func (v *connProcessor) updateDeadline() error {
	return errors.Wrap(
		v.conn.SetDeadline(time.Now().Add(v.timeout)),
		v.conn.SetReadDeadline(time.Now().Add(v.timeout)),
		v.conn.SetWriteDeadline(time.Now().Add(v.timeout)),
	)
}
