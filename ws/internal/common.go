/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.osspkg.com/errors"
)

const (
	PongWait      = 60 * time.Second
	PingPeriod    = PongWait / 3
	BusBufferSize = 128
)

var (
	ErrUnknownEventID = errors.New("unknown event id")
)

func NewUpgrade() *websocket.Upgrader {
	return &websocket.Upgrader{
		EnableCompression: false,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		WriteBufferPool:   &sync.Pool{},
		CheckOrigin: func(_ *http.Request) bool {
			return true
		},
	}
}

func SetupPingPong(c *websocket.Conn) {
	c.SetPingHandler(func(_ string) error {
		return c.SetReadDeadline(time.Now().Add(PongWait))
	})
	c.SetPongHandler(func(_ string) error {
		return c.SetReadDeadline(time.Now().Add(PongWait))
	})
}
