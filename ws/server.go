/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/do"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/goppy/v2/ws/event"
	"go.osspkg.com/goppy/v2/ws/internal"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
)

type _server struct {
	clients   map[string]*connect
	events    map[event.Id]EventHandler
	upgrade   *websocket.Upgrader
	guard     Guard
	ctx       context.Context
	cancel    context.CancelFunc
	openFunc  []func(cid string)
	closeFunc []func(cid string)
	mux       syncing.Lock
	wg        syncing.Group
}

func NewServer(ctx context.Context, opts ...func(u *websocket.Upgrader)) Server {
	upgrade := internal.NewUpgrade()
	ctx, cancel := context.WithCancel(ctx)
	for _, opt := range opts {
		opt(upgrade)
	}
	return &_server{
		clients:   make(map[string]*connect, 100),
		events:    make(map[event.Id]EventHandler, 10),
		upgrade:   upgrade,
		guard:     func(_ string, _ http.Header) error { return nil },
		ctx:       ctx,
		cancel:    cancel,
		openFunc:  make([]func(cid string), 0, 2),
		closeFunc: make([]func(cid string), 0, 2),
		mux:       syncing.NewLock(),
		wg:        syncing.NewGroup(),
	}
}

func (v *_server) Up() error {
	return nil
}

func (v *_server) Down() error {
	v.CloseAll()
	return nil
}

func (v *_server) CountConn() (cc int) {
	v.mux.RLock(func() {
		cc = len(v.clients)
	})
	return
}

func (v *_server) GetEventHandler(eid event.Id) (h EventHandler, ok bool) {
	v.mux.RLock(func() {
		h, ok = v.events[eid]
	})
	return
}

func (v *_server) SetEventHandler(h EventHandler, eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			v.events[eid] = h
		}
	})
}

func (v *_server) DelEventHandler(eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			delete(v.events, eid)
		}
	})
}

func (v *_server) SetGuard(g Guard) {
	v.mux.Lock(func() {
		v.guard = g
	})
}

func (v *_server) callGuard(cid string, head http.Header) (err error) {
	v.mux.RLock(func() {
		err = v.guard(cid, head)
	})
	return
}

func (v *_server) BroadcastEvent(eid event.Id, m interface{}) (err error) {
	event.New(func(ev event.Event) {
		ev.WithID(eid)
		if err = ev.Encode(m); err != nil {
			logx.Error("WS Server", "do", "broadcast event", "err", err, "eid", eid)
			return
		}

		var b []byte
		b, err = json.Marshal(ev)
		if err != nil {
			logx.Error("WS Server", "do", "broadcast event", "err", err, "eid", eid)
			return
		}

		v.mux.RLock(func() {
			for _, c := range v.clients {
				c.AppendMessage(b)
			}
		})
	})
	return
}

func (v *_server) SendEvent(eid event.Id, m interface{}, cids ...string) (err error) {
	event.New(func(ev event.Event) {
		ev.WithID(eid)
		if err = ev.Encode(m); err != nil {
			logx.Error("WS Server", "do", "send event", "err", err, "eid", eid)
			return
		}

		var b []byte
		b, err = json.Marshal(ev)
		if err != nil {
			logx.Error("WS Server", "do", "send event", "err", err, "eid", eid)
			return
		}

		v.mux.RLock(func() {
			for _, cid := range cids {
				c, ok := v.clients[cid]
				if !ok {
					continue
				}
				c.AppendMessage(b)
			}
		})
	})
	return
}

func (v *_server) addConn(c *connect) {
	v.mux.Lock(func() {
		v.clients[c.ConnectID()] = c
	})

	go v.mux.RLock(func() {
		for _, cb := range v.openFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(c.ConnectID())
				})
				if err != nil {
					logx.Error("WS Server", "do", "run open func", "panic", err, "cid", c.ConnectID())
				}
			}(cb)
		}
	})
}

func (v *_server) delConn(id string) {
	v.mux.Lock(func() {
		delete(v.clients, id)
	})

	go v.mux.RLock(func() {
		for _, cb := range v.closeFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(id)
				})
				if err != nil {
					logx.Error("WS Server", "do", "run close func", "panic", err, "cid", id)
				}
			}(cb)
		}
	})
}

func (v *_server) OnClose(cb func(cid string)) {
	v.mux.Lock(func() {
		v.closeFunc = append(v.closeFunc, cb)
	})
}

func (v *_server) OnOpen(cb func(cid string)) {
	v.mux.Lock(func() {
		v.openFunc = append(v.openFunc, cb)
	})
}

func (v *_server) CloseOne(cid string) {
	var conn *connect
	v.mux.RLock(func() {
		if cc, ok := v.clients[cid]; ok {
			conn = cc
		}
	})

	if conn == nil {
		return
	}

	conn.Close()
}

func (v *_server) CloseAll() {
	v.cancel()
	v.wg.Wait()
}

func (v *_server) Handling(ctx web.Context) {
	v.HandlingHTTP(ctx.Response(), ctx.Request())
}

func (v *_server) HandlingHTTP(w http.ResponseWriter, r *http.Request) {
	v.wg.Run(func() {
		cid := r.Header.Get("Sec-Websocket-Key")

		defer func() {
			if err := r.Body.Close(); err != nil {
				logx.Error("WS Server", "do", "close connect body", "err", err, "cid", cid)
			}
		}()

		conn, err := v.upgrade.Upgrade(w, r, nil)
		if err != nil {
			logx.Error("WS Server", "do", "handling: upgrade new connect", "err", err, "cid", cid)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = v.callGuard(cid, r.Header); err != nil {
			logx.Error("WS Server", "do", "handling: guard", "err", err, "cid", cid)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ctx, cancel := xc.Join(v.ctx, r.Context())
		defer cancel()

		c := newConnect(ctx, cid, r.Header, v, conn)

		c.OnClose(func(cid string) {
			v.delConn(cid)
		})
		c.OnOpen(func(string) {
			v.addConn(c)
		})
		c.Run()
	})
}
