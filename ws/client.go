/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/do"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2/ws/event"
)

type (
	Option interface {
		Set(key, value string)
	}
)

type _client struct {
	events    map[event.Id]EventHandler
	servers   map[string]*connect
	ctx       context.Context
	cancel    context.CancelFunc
	openFunc  []func(cid string)
	closeFunc []func(cid string)
	mux       syncing.Lock
	wg        syncing.Group
}

func NewClient(ctx context.Context) Client {
	ctx, cancel := context.WithCancel(ctx)
	return &_client{
		events:    make(map[event.Id]EventHandler, 10),
		servers:   make(map[string]*connect, 10),
		ctx:       ctx,
		cancel:    cancel,
		openFunc:  make([]func(cid string), 0, 2),
		closeFunc: make([]func(cid string), 0, 2),
		mux:       syncing.NewLock(),
		wg:        syncing.NewGroup(),
	}
}

func (v *_client) Up() error {
	return nil
}

func (v *_client) Down() error {
	v.CloseAll()
	return nil
}

func (v *_client) Open(url string, opts ...func(Option)) (string, error) {
	headers := make(http.Header)
	for _, opt := range opts {
		opt(headers)
	}

	conn, resp, err := websocket.DefaultDialer.DialContext(v.ctx, url, headers)
	if err != nil {
		logx.Error("WS Client", "do", "open connect", "err", err, "url", url)
		return "", err
	}

	cid := resp.Header.Get("Sec-WebSocket-Accept")

	v.wg.Background(func() {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logx.Error("WS Client", "do", "close connect body", "err", err, "url", url)
				return
			}
		}()

		c := newConnect(v.ctx, cid, resp.Header, v, conn)

		c.OnClose(func(cid string) {
			v.delConn(cid)
		})
		c.OnOpen(func(string) {
			v.addConn(c)
		})
		c.Run()
	})

	return cid, nil
}

func (v *_client) GetEventHandler(eid event.Id) (h EventHandler, ok bool) {
	v.mux.RLock(func() {
		h, ok = v.events[eid]
	})
	return
}

func (v *_client) SetEventHandler(h EventHandler, eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			v.events[eid] = h
		}
	})
}

func (v *_client) DelEventHandler(eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			delete(v.events, eid)
		}
	})
}

func (v *_client) CountConn() (cc int) {
	v.mux.RLock(func() {
		cc = len(v.servers)
	})
	return
}

func (v *_client) addConn(c *connect) {
	v.mux.Lock(func() {
		v.servers[c.ConnectID()] = c
	})

	go v.mux.RLock(func() {
		for _, cb := range v.openFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(c.ConnectID())
				})
				if err != nil {
					logx.Error("WS Client", "do", "run open func", "panic", err, "cid", c.ConnectID())
				}
			}(cb)
		}
	})
}

func (v *_client) delConn(id string) {
	v.mux.Lock(func() {
		delete(v.servers, id)
	})

	go v.mux.RLock(func() {
		for _, cb := range v.closeFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(id)
				})
				if err != nil {
					logx.Error("WS Client", "do", "run close func", "panic", err, "cid", id)
				}
			}(cb)
		}
	})
}

func (v *_client) OnClose(cb func(cid string)) {
	v.mux.Lock(func() {
		v.closeFunc = append(v.closeFunc, cb)
	})
}

func (v *_client) OnOpen(cb func(cid string)) {
	v.mux.Lock(func() {
		v.openFunc = append(v.openFunc, cb)
	})
}

func (v *_client) CloseOne(cid string) {
	var conn *connect
	v.mux.RLock(func() {
		if cc, ok := v.servers[cid]; ok {
			conn = cc
		}
	})

	if conn == nil {
		return
	}

	conn.Close()
}

func (v *_client) CloseAll() {
	v.cancel()
	v.wg.Wait()
}

func (v *_client) BroadcastEvent(eid event.Id, m interface{}) (err error) {
	event.New(func(ev event.Event) {
		ev.WithID(eid)
		if err = ev.Encode(m); err != nil {
			logx.Error("WS Client", "do", "broadcast event", "err", err, "eid", eid)
			return
		}

		var b []byte
		b, err = json.Marshal(ev)
		if err != nil {
			logx.Error("WS Client", "do", "broadcast event", "err", err, "eid", eid)
			return
		}

		v.mux.RLock(func() {
			for _, c := range v.servers {
				c.AppendMessage(b)
			}
		})
	})
	return
}

func (v *_client) SendEvent(eid event.Id, m interface{}, cids ...string) (err error) {
	event.New(func(ev event.Event) {
		ev.WithID(eid)
		if err = ev.Encode(m); err != nil {
			logx.Error("WS Client", "do", "send event", "err", err, "eid", eid)
			return
		}

		var b []byte
		b, err = json.Marshal(ev)
		if err != nil {
			logx.Error("WS Client", "do", "send event", "err", err, "eid", eid)
			return
		}

		v.mux.RLock(func() {
			for _, cid := range cids {
				c, ok := v.servers[cid]
				if !ok {
					continue
				}
				c.AppendMessage(b)
			}
		})
	})
	return
}
