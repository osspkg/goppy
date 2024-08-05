/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
)

type (
	Ctx interface {
		ConnectID() string
		WriteMessage(b []byte)
		ReadMessage() <-chan []byte
		Connect() *websocket.Conn
		Done() <-chan struct{}
		Close()
	}
)

func IsClosingError(err error) bool {
	if err == nil {
		return false
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		errors.Is(err, websocket.ErrCloseSent) {
		return true
	}
	return false
}

func PumpRead(ctx Ctx) {
	defer func() {
		ctx.Close()
	}()
	for {
		_, message, err := ctx.Connect().ReadMessage()
		if err != nil {
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "read message", "err", err, "cid", ctx.ConnectID())
			}
			return
		}

		go ctx.WriteMessage(message)
	}
}

func PumpWrite(ctx Ctx) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		ctx.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			err := ctx.Connect().WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye bye!"), time.Now().Add(PongWait))
			if err != nil && !IsClosingError(err) {
				logx.Error("WS Server", "do", "close message", "err", err, "cid", ctx.ConnectID())
			}
			return

		case <-ticker.C:
			err := ctx.Connect().WriteControl(websocket.PingMessage, nil, time.Now().Add(PongWait))
			if err == nil {
				continue
			}
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "send ping", "err", err, "cid", ctx.ConnectID())
			}
			return

		case m := <-ctx.ReadMessage():
			err := ctx.Connect().WriteMessage(websocket.TextMessage, m)
			if err == nil {
				continue
			}
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "write message", "err", err, "cid", ctx.ConnectID())
			}
			return

		}
	}
}
