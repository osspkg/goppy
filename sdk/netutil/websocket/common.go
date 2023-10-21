/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

import (
	"net/http"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/osspkg/goppy/sdk/errors"
)

const (
	pongWait      = 60 * time.Second
	pingPeriod    = pongWait / 3
	busBufferSize = 128
)

var (
	errOneOpenConnect = errors.New("connection can be started once")
	errUnknownEventID = errors.New("unknown event id")
)

func newUpgrader() ws.Upgrader {
	return ws.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
}

func setupPingPong(c *ws.Conn) {
	c.SetPingHandler(func(_ string) error {
		return errors.Wrap(
			c.SetReadDeadline(time.Now().Add(pongWait)),
			//v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})
	c.SetPongHandler(func(_ string) error {
		return errors.Wrap(
			c.SetReadDeadline(time.Now().Add(pongWait)),
			//v.conn.SetWriteDeadline(time.Now().Add(pongWait)),
		)
	})
}
