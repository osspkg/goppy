/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"

	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
	"go.osspkg.com/goppy/sdk/netutil/websocket"
)

func WithWebsocketClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l log.Logger) (*wscProvider, WebsocketClient) {
			c := newWSClientProvider(l)
			return c, c
		},
	}
}

type (
	wscProvider struct {
		connects map[string]websocket.Client

		cancel context.CancelFunc
		ctx    context.Context

		sync iosync.Switch
		mux  iosync.Lock
		wg   iosync.Group

		log log.Logger
	}

	WebsocketClient interface {
		Create(url string, opts ...func(websocket.ClientOption)) WebsocketClientConnect
	}

	WebsocketClientConnect interface {
		SendEvent(eid websocket.EventID, in interface{})
		ConnectID() string
		Header(key, value string)
		SetHandler(call websocket.ClientHandler, eids ...websocket.EventID)
		DelHandler(eids ...websocket.EventID)
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
		Close()
	}
)

func newWSClientProvider(l log.Logger) *wscProvider {
	return &wscProvider{
		connects: make(map[string]websocket.Client, 2),
		sync:     iosync.NewSwitch(),
		log:      l,
		mux:      iosync.NewLock(),
		wg:       iosync.NewGroup(),
	}
}

func (v *wscProvider) Up(ctx app.Context) error {
	if v.sync.On() {
		v.ctx, v.cancel = context.WithCancel(ctx.Context())
	}
	return nil
}

func (v *wscProvider) Down() error {
	if !v.sync.Off() {
		return nil
	}
	v.cancel()
	v.wg.Wait()
	return nil
}

func (v *wscProvider) addConn(cc websocket.Client) {
	v.mux.Lock(func() {
		v.connects[cc.ConnectID()] = cc
	})
}

func (v *wscProvider) delConn(cid string) {
	v.mux.Lock(func() {
		delete(v.connects, cid)
	})
}

func (v *wscProvider) errLog(cid string, err error, msg string, args ...interface{}) {
	if err == nil || v.log == nil {
		return
	}
	v.log.WithFields(log.Fields{
		"cid": cid,
		"err": err.Error(),
	}).Errorf(msg, args...)
}

func (v *wscProvider) Create(url string, opts ...func(websocket.ClientOption)) WebsocketClientConnect {
	cc := websocket.NewClient(v.ctx, url, v.log, opts...)

	cc.OnClose(func(cid string) {
		v.delConn(cid)
	})
	cc.OnOpen(func(cid string) {
		v.addConn(cc)
	})

	v.wg.Background(func() {
		if err := cc.DialAndListen(); err != nil {
			v.errLog(cc.ConnectID(), err, "[ws] dial to %s", url)
		}
	})
	return cc
}
