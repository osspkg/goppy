/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unixsocket

import (
	"io"

	"github.com/osspkg/goppy/sdk/errors"
	"github.com/osspkg/goppy/sdk/ioutil"
)

var (
	newLine    = "\n"
	divideStr  = " "
	divideByte = byte(' ')

	ErrInvalidCommand = errors.New("command not found")
)

func writeError(v io.Writer, err error) error {
	return ioutil.WriteBytes(v, []byte(err.Error()), newLine)
}

func parseCommand(b []byte) (string, []byte) {
	for i := 0; i < len(b); i++ {
		if b[i] == divideByte {
			if len(b) > i+2 {
				return string(b[0:i]), b[i+1:]
			}
			return string(b[0:i]), nil
		}
	}
	return string(b), nil
}
