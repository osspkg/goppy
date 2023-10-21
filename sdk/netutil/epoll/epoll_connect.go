/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package epoll

import (
	"bytes"
	"io"
	"sync"

	"github.com/osspkg/goppy/sdk/errors"
)

var (
	epollBodyPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}

	errInvalidPoolType = errors.New("invalid data type from pool")
)

type Handler func([]byte, io.Writer) error

func newEpollConn(conn io.ReadWriter, handler Handler, eof []byte) error {
	var (
		n   int
		err error
		l   = len(eof)
	)
	b, ok := epollBodyPool.Get().([]byte)
	if !ok {
		return errInvalidPoolType
	}
	defer epollBodyPool.Put(b[:0]) //nolint:staticcheck

	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err = conn.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if len(b) < l {
			return io.EOF
		}
		if bytes.Equal(eof, b[len(b)-l:]) {
			b = b[:len(b)-l]
			break
		}
	}
	err = handler(b, conn)
	return err
}
