/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unixsocket

import (
	"net"

	"github.com/osspkg/goppy/sdk/errors"
	"github.com/osspkg/goppy/sdk/ioutil"
)

type Client struct {
	path string
}

func NewClient(path string) *Client {
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
	if err = ioutil.WriteBytes(conn, append([]byte(name+divideStr), b...), newLine); err != nil {
		return nil, err
	}
	return ioutil.ReadBytes(conn, newLine)
}

func (v *Client) ExecString(name string, b string) ([]byte, error) {
	return v.Exec(name, []byte(b))
}
