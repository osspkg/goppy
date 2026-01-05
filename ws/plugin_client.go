/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
	"go.osspkg.com/goppy/v3/ws/event"
)

type Client interface {
	Open(url string, opts ...func(h http.Header, d *websocket.Dialer)) (string, error)
	SetEventHandler(handler EventHandler, eventIDs ...event.Id)
	DelEventHandler(eventIDs ...event.Id)
	CountConn() int
	AddOnCloseFunc(callback func(serverId string))
	AddOnOpenFunc(callback func(serverId string))
	CloseConnect(serverId string)
	CloseAll()
	BroadcastEvent(eventId event.Id, message any) error
	SendEvent(eventId event.Id, message any, serverIDs ...string) error
}

func WithClient() plugins.Kind {
	return plugins.Kind{
		Inject: func(ctx xc.Context) Client {
			return NewClient(ctx.Context())
		},
	}
}
