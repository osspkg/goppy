/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"net"
)

type HandlerTCP interface {
	HandlerTCP(p Processor)
}

type Processor interface {
	Write([]byte) (int, error)
	ReadLine() ([]byte, error)
	Read(b []byte) (int, error)
	Addr() net.Addr
}
