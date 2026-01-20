/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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
	connect interface {
		ConnectID() string
		ReceiveMessage(b []byte)
		SendMessageChan() <-chan []byte
		Connect() *websocket.Conn
		Done() <-chan struct{}
		Close()
	}
)

func IsClosingError(err error) bool {
	if err == nil {
		return false
	}

	switch {
	case errors.Is(err, websocket.ErrCloseSent):
		return true
	case websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived):
		return true
	case strings.Contains(err.Error(), "connection reset by peer"):
		return true
	case strings.Contains(err.Error(), "use of closed network connection"):
		return true
	default:
		return false
	}
}

func PumpRead(cc connect) {
	defer func() {
		cc.Close()
	}()

	for {
		_, message, err := cc.Connect().ReadMessage()
		if err != nil {
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "read message", "err", err, "cid", cc.ConnectID())
			}
			return
		}

		go cc.ReceiveMessage(message)
	}
}

func PumpWrite(cc connect) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		cc.Close()
	}()

	closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye bye!")

	for {
		select {
		case <-cc.Done():
			err := cc.Connect().WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(PongWait))
			if err != nil && !IsClosingError(err) {
				logx.Error("WS Server", "do", "close message", "err", err, "cid", cc.ConnectID())
			}
			return

		case <-ticker.C:
			err := cc.Connect().WriteControl(websocket.PingMessage, nil, time.Now().Add(PongWait))
			if err == nil {
				continue
			}
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "send ping", "err", err, "cid", cc.ConnectID())
			}
			return

		case message := <-cc.SendMessageChan():
			err := cc.Connect().WriteMessage(websocket.TextMessage, message)
			if err == nil {
				continue
			}
			if !IsClosingError(err) {
				logx.Error("WS Server", "do", "write message", "err", err, "cid", cc.ConnectID())
			}
			return

		}
	}
}
