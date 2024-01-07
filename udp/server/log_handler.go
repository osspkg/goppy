/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"net"

	"go.osspkg.com/goppy/xlog"
)

type logHandler struct {
	log xlog.Logger
}

func NewLogHandlerUDP(l xlog.Logger) HandlerUDP {
	return &logHandler{log: l}
}

func (v *logHandler) HandlerUDP(_ Writer, addr net.Addr, b []byte) {
	v.log.WithFields(xlog.Fields{
		"addr": addr.String(),
		"body": string(b),
	}).Warnf("Empty log handler UDP")
}
