/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"net"
)

type HandlerTCP interface {
	HandlerTCP(w Response, r Request)
}

type (
	Request interface {
		ReadLine() ([]byte, error)
		Read(b []byte) (int, error)
		Addr() net.Addr
	}

	Response interface {
		Write([]byte) (int, error)
	}
)
