/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/errors"
)

const (
	PongWait      = 60 * time.Second
	PingPeriod    = PongWait / 3
	BusBufferSize = 128
)

var (
	ErrOneOpenConnect = errors.New("connection can be started once")
	ErrUnknownEventID = errors.New("unknown event id")
)

func NewUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
}

func SetupPingPong(c *websocket.Conn) {
	c.SetPingHandler(func(_ string) error {
		return errors.Wrap(
			c.SetReadDeadline(time.Now().Add(PongWait)),
			//v.conn.SetWriteDeadline(time.Now().Add(PongWait)),
		)
	})
	c.SetPongHandler(func(_ string) error {
		return errors.Wrap(
			c.SetReadDeadline(time.Now().Add(PongWait)),
			//v.conn.SetWriteDeadline(time.Now().Add(PongWait)),
		)
	})
}
