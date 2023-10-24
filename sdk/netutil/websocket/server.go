/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
)

type Server struct {
	clients map[string]*Connect
	events  map[EventID]EventHandler
	upgrade websocket.Upgrader
	logger  log.Logger
	ctx     context.Context
	cancel  context.CancelFunc
	mux     iosync.Lock
	wg      iosync.Group
}

func NewServer(l log.Logger, ctx context.Context, opts ...func(u websocket.Upgrader)) *Server {
	up := newUpgrader()
	c, cancel := context.WithCancel(ctx)
	for _, opt := range opts {
		opt(up)
	}
	return &Server{
		clients: make(map[string]*Connect, 100),
		events:  make(map[EventID]EventHandler, 10),
		upgrade: up,
		logger:  l,
		ctx:     c,
		cancel:  cancel,
		mux:     iosync.NewLock(),
		wg:      iosync.NewGroup(),
	}
}

func (v *Server) CloseAll() {
	v.cancel()
	v.wg.Wait()
}

func (v *Server) ErrLog(cid string, err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}
	v.logger.WithFields(log.Fields{"cid": cid, "err": err.Error()}).Errorf(msg, args...)
}

func (v *Server) ErrLogMessage(cid string, msg string, args ...interface{}) {
	v.logger.WithFields(log.Fields{"cid": cid}).Errorf(msg, args...)
}

func (v *Server) CountConn() (cc int) {
	v.mux.Lock(func() {
		cc = len(v.clients)
	})
	return
}

func (v *Server) AddConn(c *Connect) {
	v.mux.Lock(func() {
		v.clients[c.ConnectID()] = c
	})
}

func (v *Server) DelConn(id string) {
	v.mux.Lock(func() {
		delete(v.clients, id)
	})
}

func (v *Server) SetHandler(call EventHandler, eids ...EventID) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			v.events[eid] = call
		}
	})
}

func (v *Server) GetHandler(eid EventID) (h EventHandler, ok bool) {
	v.mux.RLock(func() {
		h, ok = v.events[eid]
	})
	return
}

func (v *Server) Broadcast(eid EventID, m interface{}) {
	getEventModel(func(ev *event) {
		ev.ID = eid
		b, err := json.Marshal(m)
		if err != nil {
			v.ErrLog("*", err, "[ws] broadcast error")
			return
		}
		ev.Body(b)

		b, err = json.Marshal(ev)
		if err != nil {
			v.ErrLog("*", err, "[ws] broadcast error")
			return
		}
		v.mux.RLock(func() {
			for _, c := range v.clients {
				c.WriteToBus(b)
			}
		})
	})
}

func (v *Server) SendEvent(eid EventID, m interface{}, cids ...string) {
	getEventModel(func(ev *event) {
		ev.ID = eid
		b, err := json.Marshal(m)
		if err != nil {
			v.ErrLog("*", err, "[ws] send event error")
			return
		}
		ev.Body(b)
		b, err = json.Marshal(ev)
		if err != nil {
			v.ErrLog("*", err, "[ws] send event error")
			return
		}
		v.mux.RLock(func() {
			for _, cid := range cids {
				if c, ok := v.clients[cid]; ok {
					c.WriteToBus(b)
				}
			}
		})
	})
}

func (v *Server) Handling(w http.ResponseWriter, r *http.Request) {
	v.wg.Run(func() {
		cid := r.Header.Get("Sec-Websocket-Key")
		up, err := v.upgrade.Upgrade(w, r, nil)
		if err != nil {
			v.ErrLog(cid, err, "[ws] upgrade")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		c := NewConnect(cid, r.Header, v, up, r.Context(), v.ctx)
		c.OnClose(func(cid string) {
			v.DelConn(cid)
		})
		c.OnOpen(func(string) {
			v.AddConn(c)
		})
		c.Run()
	})
}
