/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ioutil

import (
	"bytes"
	"io"

	"go.osspkg.com/goppy/sdk/errors"
)

func ReadAll(r io.ReadCloser) ([]byte, error) {
	b, err := io.ReadAll(r)
	err = errors.Wrap(err, r.Close())
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ReadBytes(v io.Reader, divide string) ([]byte, error) {
	var (
		n   int
		err error
		b   = make([]byte, 0, 512)
		db  = []byte(divide)
		dl  = len(db)
	)

	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err = v.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(b) < dl {
			return b, io.EOF
		}
		if bytes.Equal(db, b[len(b)-dl:]) {
			b = b[:len(b)-dl]
			break
		}
	}
	return b, nil
}

func WriteBytes(v io.Writer, b []byte, divide string) error {
	var (
		db = []byte(divide)
		dl = len(db)
	)
	if len(b) < dl || !bytes.Equal(db, b[len(b)-dl:]) {
		b = append(b, db...)
	}
	if _, err := v.Write(b); err != nil {
		return err
	}
	return nil
}
