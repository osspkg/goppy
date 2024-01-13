/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/ws/server"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

func WebsocketServerOptionCompression(enable bool) func(websocket.Upgrader) {
	return func(upg websocket.Upgrader) {
		upg.EnableCompression = enable
	}
}

func WebsocketServerOptionBuffer(read, write int) func(websocket.Upgrader) {
	return func(upg websocket.Upgrader) {
		upg.ReadBufferSize, upg.WriteBufferSize = read, write
	}
}

func WithWebsocketServer(options ...func(websocket.Upgrader)) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l xlog.Logger, ctx xc.Context) (*wssProvider, WebsocketServer) {
			wsp := newWsServerProvider(l, ctx, options...)
			return wsp, wsp.serv
		},
	}
}

type (
	WebsocketServer interface {
		Handling(w http.ResponseWriter, r *http.Request)
		SendEvent(eid event.Id, m interface{}, cids ...string)
		Broadcast(eid event.Id, m interface{})
		SetHandler(call server.EventHandler, eids ...event.Id)
		CloseAll()
		CountConn() int
	}

	wssProvider struct {
		log  xlog.Logger
		serv *server.Server
		sync iosync.Switch
	}
)

func newWsServerProvider(l xlog.Logger, ctx xc.Context, options ...func(websocket.Upgrader)) *wssProvider {
	return &wssProvider{
		log:  l,
		serv: server.New(l, ctx.Context(), options...),
		sync: iosync.NewSwitch(),
	}
}

func (v *wssProvider) Up() error {
	if !v.sync.On() {
		return errServAlreadyRunning
	}
	v.log.Infof("Websocket started")
	return nil
}

func (v *wssProvider) Down() error {
	if !v.sync.Off() {
		return errServAlreadyStopped
	}
	v.serv.CloseAll()
	v.log.Infof("Websocket stopped")
	return nil
}
