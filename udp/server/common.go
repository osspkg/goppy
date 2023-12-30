/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"net"
)

type (
	HandlerUDP interface {
		HandlerUDP(w Writer, addr net.Addr, b []byte)
	}
	Writer interface {
		WriteTo(p []byte, addr net.Addr) (n int, err error)
	}
)
