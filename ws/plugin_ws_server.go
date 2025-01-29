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
	CountConn() (cc int)
	SetEventHandler(h EventHandler, eids ...event.Id)
	DelEventHandler(eids ...event.Id)
	SetGuard(g Guard)
	BroadcastEvent(eid event.Id, m interface{}) (err error)
	SendEvent(eid event.Id, m interface{}, cids ...string) (err error)
	OnClose(cb func(cid string))
	OnOpen(cb func(cid string))
	CloseOne(cid string)
	CloseAll()
	HandlingHTTP(w http.ResponseWriter, r *http.Request)
	Handling(ctx web.Context)
}

func OptionCompression(enable bool) func(*websocket.Upgrader) {
	return func(u *websocket.Upgrader) {
		u.EnableCompression = enable
	}
}

func OptionBuffer(read, write int) func(*websocket.Upgrader) {
	return func(u *websocket.Upgrader) {
		u.ReadBufferSize, u.WriteBufferSize = read, write
	}
}

func WithServer(options ...func(*websocket.Upgrader)) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(ctx xc.Context) Server {
			return NewServer(ctx.Context(), options...)
		},
	}
}
