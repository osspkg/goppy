/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"io"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/ioutil"
)

var (
	NewLine    = "\n"
	DivideStr  = " "
	DivideByte = byte(' ')

	ErrInvalidCommand = errors.New("command not found")
)

func WriteError(v io.Writer, err error) error {
	return ioutil.WriteBytes(v, []byte(err.Error()), NewLine)
}

func ParseCommand(b []byte) (string, []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == DivideByte {
			if len(b) > i+2 {
				return string(b[0:i]), b[i+1:]
			}
			return string(b[0:i]), nil
		}
	}
	return string(b), nil
}
