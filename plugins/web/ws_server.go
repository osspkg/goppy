/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"net/http"

	ws "github.com/gorilla/websocket"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
	"go.osspkg.com/goppy/sdk/netutil/websocket"
)

func WebsocketServerOptionCompression(enable bool) func(ws.Upgrader) {
	return func(upg ws.Upgrader) {
		upg.EnableCompression = enable
	}
}

func WebsocketServerOptionBuffer(read, write int) func(ws.Upgrader) {
	return func(upg ws.Upgrader) {
		upg.ReadBufferSize, upg.WriteBufferSize = read, write
	}
}

func WithWebsocketServer(options ...func(ws.Upgrader)) plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l log.Logger, ctx app.Context) (*wssProvider, WebsocketServer) {
			wsp := newWsServerProvider(l, ctx, options...)
			return wsp, wsp.serv
		},
	}
}

type (
	WebsocketServer interface {
		Handling(w http.ResponseWriter, r *http.Request)
		SendEvent(eid websocket.EventID, m interface{}, cids ...string)
		Broadcast(eid websocket.EventID, m interface{})
		SetHandler(call websocket.EventHandler, eids ...websocket.EventID)
		CloseAll()
		CountConn() int
	}

	wssProvider struct {
		log  log.Logger
		serv *websocket.Server
		sync iosync.Switch
	}
)

func newWsServerProvider(l log.Logger, ctx app.Context, options ...func(ws.Upgrader)) *wssProvider {
	return &wssProvider{
		log:  l,
		serv: websocket.NewServer(l, ctx.Context(), options...),
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
