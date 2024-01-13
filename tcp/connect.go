/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/ioutil"
	"go.osspkg.com/goppy/random"
	"go.osspkg.com/goppy/xc"
)

var randomBytes = []byte("1234567890qazwsxedcrfvtgbyhnujmikolpQAZWSXEDCRFVTGBYHNUJMIKOLP")

type (
	Connect interface {
		io.ReadWriter
		Addr() string
		ID() string
		Close() error
		Closed() <-chan struct{}
	}

	ErrConnect interface {
		io.Writer
		Addr() string
		ID() string
		Err() error
	}

	connectProvider struct {
		id   string
		conf cpConfig
		conn net.Conn
		buff *bytes.Buffer
		ctx  xc.Context
		err  error
	}

	cpConfig struct {
		MaxSize int
		Timeout time.Duration
	}
)

func newConnectProvider(ctx context.Context, c net.Conn, cpc cpConfig) *connectProvider {
	return &connectProvider{
		id:   string(random.BytesOf(32, randomBytes)),
		buff: bytes.NewBuffer(nil),
		conn: c,
		ctx:  xc.NewContext(ctx),
		conf: cpc,
	}
}

func (v *connectProvider) updateDeadline(c net.Conn) error {
	t := time.Now().Add(v.conf.Timeout)
	return errors.Wrap(
		c.SetDeadline(t),
		c.SetReadDeadline(t),
		c.SetWriteDeadline(t),
	)
}

func (v *connectProvider) Wait() error {
	if v.err = v.updateDeadline(v.conn); v.err != nil {
		return v.err
	}
	select {
	case <-v.ctx.Done():
		return io.EOF
	default:
	}
	v.buff.Reset()
	if v.err = ioutil.ReadFull(v.buff, v.conn, v.conf.MaxSize); v.err != nil {
		return v.err
	}
	return nil
}

func (v *connectProvider) IsEmpty() bool {
	return v.buff.Len() == 0
}

func (v *connectProvider) Read(b []byte) (int, error) {
	return v.buff.Read(b)
}

func (v *connectProvider) Addr() string {
	return v.conn.RemoteAddr().String()
}

func (v *connectProvider) Write(b []byte) (int, error) {
	return v.conn.Write(b)
}

func (v *connectProvider) ID() string {
	return v.id
}

func (v *connectProvider) Err() error {
	err := v.err
	v.err = nil
	return err
}

func (v *connectProvider) Close() error {
	v.ctx.Close()
	return nil
}

func (v *connectProvider) Closed() <-chan struct{} {
	return v.ctx.Done()
}
