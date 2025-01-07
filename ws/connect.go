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
	"go.osspkg.com/goppy/v2/ws/event"
	"go.osspkg.com/goppy/v2/ws/internal"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
)

type (
	resolver interface {
		GetEventHandler(eid event.Id) (EventHandler, bool)
	}

	connect struct {
		id        string
		header    http.Header
		resolver  resolver
		conn      *websocket.Conn
		dataC     chan []byte
		ctx       context.Context
		cancel    context.CancelFunc
		openFunc  []func(cid string)
		closeFunc []func(cid string)
		status    syncing.Switch
		mux       syncing.Lock
	}
)

func newConnect(ctx context.Context, id string, head http.Header, r resolver, conn *websocket.Conn) *connect {
	ctx, cancel := context.WithCancel(ctx)
	return &connect{
		id:        id,
		header:    head,
		resolver:  r,
		conn:      conn,
		dataC:     make(chan []byte, internal.BusBufferSize),
		ctx:       ctx,
		cancel:    cancel,
		closeFunc: make([]func(string), 0, 2),
		openFunc:  make([]func(string), 0, 2),
		status:    syncing.NewSwitch(),
		mux:       syncing.NewLock(),
	}
}

func (v *connect) Context() context.Context {
	return v.ctx
}

func (v *connect) ConnectID() string {
	return v.id
}

func (v *connect) Head(key string) string {
	return v.header.Get(key)
}

func (v *connect) Connect() *websocket.Conn {
	return v.conn
}

func (v *connect) Done() <-chan struct{} {
	return v.ctx.Done()
}

func (v *connect) Run() {
	if !v.status.On() {
		return
	}

	v.mux.RLock(func() {
		for _, cb := range v.openFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(v.ConnectID())
				})
				if err != nil {
					logx.Error("WS Connect", "do", "run open func", "panic", err, "cid", v.ConnectID())
				}
			}(cb)
		}
	})

	internal.SetupPingPong(v.conn)

	wg := syncing.NewGroup()
	wg.Background(func() { internal.PumpWrite(v) })
	wg.Background(func() { internal.PumpRead(v) })
	wg.Wait()

	v.mux.RLock(func() {
		for _, cb := range v.closeFunc {
			go func(call func(cid string)) {
				err := do.Recovery(func() {
					call(v.ConnectID())
				})
				if err != nil {
					logx.Error("WS Connect", "do", "run close func", "panic", err, "cid", v.ConnectID())
				}
			}(cb)
		}
	})
}

func (v *connect) Close() {
	if !v.status.Off() {
		return
	}
	v.cancel()
	if err := v.conn.Close(); err != nil && !internal.IsClosingError(err) {
		logx.Error("WS Connect", "do", "close connect", "err", err, "cid", v.ConnectID())
	}
}

func (v *connect) OnClose(cb func(cid string)) {
	v.mux.Lock(func() {
		v.closeFunc = append(v.closeFunc, cb)
	})
}

func (v *connect) OnOpen(cb func(cid string)) {
	v.mux.Lock(func() {
		v.openFunc = append(v.openFunc, cb)
	})
}

func (v *connect) ReadMessage() <-chan []byte {
	return v.dataC
}

func (v *connect) WriteMessage(b []byte) {
	event.New(func(ev event.Event) {
		if err := json.Unmarshal(b, ev); err != nil {
			logx.Error("WS Connect", "do", "decode message", "err", err, "cid", v.ConnectID())
			return
		}
		call, ok := v.resolver.GetEventHandler(ev.ID())
		if !ok {
			ev.WithError(internal.ErrUnknownEventID)
		} else if err := call(ev, v); err != nil {
			ev.WithError(err)
		}
		if bb, err := json.Marshal(ev); err != nil {
			logx.Error("WS Connect", "do", "encode message", "err", err, "cid", v.ConnectID())
		} else {
			v.AppendMessage(bb)
		}
	})
}

func (v *connect) AppendMessage(b []byte) {
	if len(b) == 0 {
		return
	}
	select {
	case v.dataC <- b:
	default:
		logx.Error("WS Connect", "do", "append message", "err", "write chan is full", "cid", v.ConnectID())
	}
	return
}

func (v *connect) Encode(eid event.Id, in interface{}) {
	event.New(func(ev event.Event) {
		ev.WithID(eid)
		if err := ev.Encode(in); err != nil {
			logx.Error("WS Connect", "do", "encode message", "err", err, "cid", v.ConnectID())
			return
		}
		b, err := json.Marshal(ev)
		if err != nil {
			logx.Error("WS Connect", "do", "encode event", "err", err, "cid", v.ConnectID())
			return
		}
		v.AppendMessage(b)
	})
}
