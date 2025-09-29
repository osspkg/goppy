/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/goppy/v2/ws/event"
)

type Server interface {
	CountConn() int
	SetEventHandler(h EventHandler, eventIDs ...event.Id)
	DelEventHandler(eventIDs ...event.Id)
	AddGuard(g Guard)
	BroadcastEvent(eventId event.Id, m any) (err error)
	SendEvent(eventId event.Id, m any, clientIDs ...string) (err error)
	AddOnCloseFunc(call func(clientId string))
	AddOnOpenFunc(cb func(clientId string))
	CloseOne(clientId string)
	CloseAll()
	HandlingHTTP(w http.ResponseWriter, r *http.Request)
	Handling(ctx web.Ctx)
}

func Compression(enable bool) func(*websocket.Upgrader) {
	return func(u *websocket.Upgrader) {
		u.EnableCompression = enable
	}
}

func ReadWriteBuffer(read, write int) func(*websocket.Upgrader) {
	return func(u *websocket.Upgrader) {
		u.ReadBufferSize, u.WriteBufferSize = read, write
	}
}

func WithServer(options ...func(*websocket.Upgrader)) plugins.Kind {
	return plugins.Kind{
		Inject: func(ctx xc.Context) Server {
			return NewServer(ctx.Context(), options...)
		},
	}
}
