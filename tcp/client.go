/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"context"
	"io"
	"net"
	"time"
)

type (
	Client struct {
		address *net.TCPAddr
		conf    cpConfig
	}

	ClientConfig struct {
		Address           string
		Timeout           time.Duration
		ServerMaxBodySize int
	}
)

func NewClient(c ClientConfig) (*Client, error) {
	hostPort, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return nil, err
	}
	cli := &Client{
		address: hostPort,
		conf: cpConfig{
			MaxSize: c.ServerMaxBodySize,
			Timeout: c.Timeout,
		},
	}
	return cli, nil
}

func (v *Client) Do(ctx context.Context, r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, v.address)
	if err != nil {
		return nil, err
	}
	defer conn.Close() // nolint: errcheck
	cp := newConnectProvider(ctx, conn, v.conf)
	if _, err = cp.Write(b); err != nil {
		return nil, err
	}
	if err = cp.Wait(); err != nil {
		return nil, err
	}
	return io.ReadAll(cp)
}
