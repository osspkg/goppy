/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"

	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/ws/client"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

func WithWebsocketClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func(l xlog.Logger) WebsocketClient {
			return newWSClientProvider(l)
		},
	}
}

type (
	wscProvider struct {
		connects map[string]client.Client

		cancel context.CancelFunc
		ctx    context.Context

		sync iosync.Switch
		mux  iosync.Lock
		wg   iosync.Group

		log xlog.Logger
	}

	WebsocketClient interface {
		Create(url string, opts ...func(option client.Option)) WebsocketClientConnect
	}

	WebsocketClientConnect interface {
		SendEvent(eid event.Id, in interface{})
		ConnectID() string
		Header(key, value string)
		SetHandler(call client.Handler, eids ...event.Id)
		DelHandler(eids ...event.Id)
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
		Close()
	}
)

func newWSClientProvider(l xlog.Logger) *wscProvider {
	return &wscProvider{
		connects: make(map[string]client.Client, 2),
		sync:     iosync.NewSwitch(),
		log:      l,
		mux:      iosync.NewLock(),
		wg:       iosync.NewGroup(),
	}
}

func (v *wscProvider) Up(ctx xc.Context) error {
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

func (v *wscProvider) addConn(cc client.Client) {
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
	v.log.WithFields(xlog.Fields{
		"cid": cid,
		"err": err.Error(),
	}).Errorf(msg, args...)
}

func (v *wscProvider) Create(url string, opts ...func(option client.Option)) WebsocketClientConnect {
	cc := client.New(v.ctx, url, v.log, opts...)

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
