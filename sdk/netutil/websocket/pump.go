/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

import (
	"context"
	"strings"
	"time"

	ws "github.com/gorilla/websocket"
	"go.osspkg.com/goppy/sdk/errors"
)

type (
	pumpApi interface {
		ConnectID() string
		callHandler(b []byte)
		readBus() <-chan []byte
		connect() *ws.Conn
		cancelFunc() context.CancelFunc
		done() <-chan struct{}
		Close()
	}
	pumpActionsApi interface {
		ErrLog(cid string, err error, msg string, args ...interface{})
	}
)

func isClosingError(err error) bool {
	if ws.IsCloseError(err, ws.CloseNormalClosure, ws.CloseGoingAway, ws.CloseNoStatusReceived) ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		errors.Is(err, ws.ErrCloseSent) {
		return true
	}
	return false
}

func pumpRead(p pumpApi, a pumpActionsApi) {
	defer p.cancelFunc()
	for {
		_, message, err := p.connect().ReadMessage()
		if err != nil {
			if !isClosingError(err) {
				a.ErrLog(p.ConnectID(), err, "[ws] read message")
			}
			return
		}
		go p.callHandler(message)
	}
}

func pumpWrite(p pumpApi, a pumpActionsApi) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		a.ErrLog(p.ConnectID(), p.connect().Close(), "close connect")
	}()
	for {
		select {
		case <-p.done():
			err := p.connect().WriteControl(ws.CloseMessage,
				ws.FormatCloseMessage(ws.CloseNormalClosure, "Bye bye!"), time.Now().Add(pongWait))
			if err != nil && !isClosingError(err) {
				a.ErrLog(p.ConnectID(), err, "[ws] send close")
			}
			return
		case <-ticker.C:
			if err := p.connect().WriteControl(ws.PingMessage, nil, time.Now().Add(pongWait)); err != nil {
				if !isClosingError(err) {
					a.ErrLog(p.ConnectID(), err, "[ws] send ping")
				}
				return
			}
		case m := <-p.readBus():
			if err := p.connect().WriteMessage(ws.TextMessage, m); err != nil {
				if !isClosingError(err) {
					a.ErrLog(p.ConnectID(), err, "[ws] send message")
				}
				return
			}
		}
	}
}
