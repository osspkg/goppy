/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package internal

import (
	"context"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/errors"
)

type (
	PumpApi interface {
		ConnectID() string
		CallHandler(b []byte)
		ReadBus() <-chan []byte
		Connect() *websocket.Conn
		CancelFunc() context.CancelFunc
		Done() <-chan struct{}
		Close()
	}
	PumpActionsApi interface {
		ErrLog(cid string, err error, msg string, args ...interface{})
	}
)

func IsClosingError(err error) bool {
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		errors.Is(err, websocket.ErrCloseSent) {
		return true
	}
	return false
}

func PumpRead(p PumpApi, a PumpActionsApi) {
	defer p.CancelFunc()
	for {
		_, message, err := p.Connect().ReadMessage()
		if err != nil {
			if !IsClosingError(err) {
				a.ErrLog(p.ConnectID(), err, "[ws] read message")
			}
			return
		}
		go p.CallHandler(message)
	}
}

func PumpWrite(p PumpApi, a PumpActionsApi) {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		a.ErrLog(p.ConnectID(), p.Connect().Close(), "close connect")
	}()
	for {
		select {
		case <-p.Done():
			err := p.Connect().WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Bye bye!"), time.Now().Add(PongWait))
			if err != nil && !IsClosingError(err) {
				a.ErrLog(p.ConnectID(), err, "[ws] send close")
			}
			return
		case <-ticker.C:
			if err := p.Connect().WriteControl(websocket.PingMessage, nil, time.Now().Add(PongWait)); err != nil {
				if !IsClosingError(err) {
					a.ErrLog(p.ConnectID(), err, "[ws] send ping")
				}
				return
			}
		case m := <-p.ReadBus():
			if err := p.Connect().WriteMessage(websocket.TextMessage, m); err != nil {
				if !IsClosingError(err) {
					a.ErrLog(p.ConnectID(), err, "[ws] send message")
				}
				return
			}
		}
	}
}
