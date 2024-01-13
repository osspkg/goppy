/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"net"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/ioutil"
	"go.osspkg.com/goppy/unixsocket/internal"
)

type Client struct {
	path string
}

func New(path string) *Client {
	return &Client{
		path: path,
	}
}

func (v *Client) Exec(name string, b []byte) ([]byte, error) {
	conn, err := net.Dial("unix", v.path)
	if err != nil {
		return nil, errors.Wrapf(err, "open connect [unix:%s]", v.path)
	}
	defer conn.Close() //nolint: errcheck
	if err = ioutil.WriteBytes(conn, append([]byte(name+internal.DivideStr), b...), internal.NewLine); err != nil {
		return nil, err
	}
	return ioutil.ReadBytes(conn, internal.NewLine)
}

func (v *Client) ExecString(name string, b string) ([]byte, error) {
	return v.Exec(name, []byte(b))
}
