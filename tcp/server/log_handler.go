/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"go.osspkg.com/goppy/xlog"
)

type logHandler struct {
	log xlog.Logger
}

func NewLogHandlerTCP(l xlog.Logger) HandlerTCP {
	return &logHandler{log: l}
}

func (v *logHandler) HandlerTCP(p Processor) {
	for {
		b, err := p.ReadLine()
		if err != nil {
			v.log.WithFields(xlog.Fields{
				"addr": p.Addr().String(),
				"err":  err.Error(),
			}).Errorf("Empty log handler TCP")
			return
		}

		v.log.WithFields(xlog.Fields{
			"addr": p.Addr().String(),
			"len":  len(b),
			"body": string(b),
		}).Warnf("Empty log handler TCP")
	}
}
