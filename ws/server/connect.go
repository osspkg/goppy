/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/ws/internal"
	context2 "go.osspkg.com/goppy/xc"
)

type (
	actionsApi interface {
		ErrLog(cid string, err error, msg string, args ...interface{})
		ErrLogMessage(cid string, msg string, args ...interface{})
		GetHandler(eid event.Id) (EventHandler, bool)
	}

	Connect struct {
		id        string
		header    http.Header
		actions   actionsApi
		conn      *websocket.Conn
		busBuf    chan []byte
		ctx       context.Context
		cancel    context.CancelFunc
		openFunc  []func(cid string)
		closeFunc []func(cid string)
		sync      iosync.Switch
		mux       iosync.Lock
	}
)

func NewConnect(
	id string, head http.Header,
	act actionsApi, conn *websocket.Conn,
	ctxs ...context.Context,
) *Connect {
	ctx, cancel := context2.Combine(ctxs...)
	return &Connect{
		id:        id,
		header:    head,
		actions:   act,
		conn:      conn,
		busBuf:    make(chan []byte, internal.BusBufferSize),
		ctx:       ctx,
		cancel:    cancel,
		closeFunc: make([]func(string), 0, 2),
		openFunc:  make([]func(string), 0, 2),
		sync:      iosync.NewSwitch(),
		mux:       iosync.NewLock(),
	}
}

func (v *Connect) ConnectID() string {
	return v.id
}

func (v *Connect) Head(key string) string {
	return v.header.Get(key)
}

func (v *Connect) Connect() *websocket.Conn {
	return v.conn
}

func (v *Connect) CancelFunc() context.CancelFunc {
	return v.cancel
}

func (v *Connect) Done() <-chan struct{} {
	return v.ctx.Done()
}

func (v *Connect) ReadBus() <-chan []byte {
	return v.busBuf
}

func (v *Connect) WriteToBus(b []byte) {
	if len(b) == 0 {
		return
	}
	select {
	case v.busBuf <- b:
	default:
		v.actions.ErrLogMessage(v.id, "write chan is full")
	}
}

func (v *Connect) Encode(eid event.Id, in interface{}) {
	event.GetMessage(func(ev *event.Message) {
		ev.ID = eid
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.actions.ErrLog(v.ConnectID(), err, "[ws] encode message: %d", eid)
			return
		}
		v.WriteToBus(b)
	})
}

func (v *Connect) CallHandler(b []byte) {
	event.GetMessage(func(ev *event.Message) {
		if err := json.Unmarshal(b, ev); err != nil {
			v.actions.ErrLog(v.ConnectID(), err, "[ws] decode message")
			return
		}
		call, ok := v.actions.GetHandler(ev.EventID())
		if !ok {
			ev.Error(internal.ErrUnknownEventID)
		} else if err := call(ev, ev, v); err != nil {
			ev.Error(err)
		}
		if bb, err := json.Marshal(ev); err != nil {
			v.actions.ErrLog(v.ConnectID(), err, "[ws] encode message: %d", ev.EventID())
		} else {
			v.WriteToBus(bb)
		}
	})
}

func (v *Connect) OnClose(cb func(cid string)) {
	v.mux.Lock(func() {
		v.closeFunc = append(v.closeFunc, cb)
	})
}

func (v *Connect) OnOpen(cb func(cid string)) {
	v.mux.Lock(func() {
		v.openFunc = append(v.openFunc, cb)
	})
}

func (v *Connect) Close() {
	if !v.sync.Off() {
		return
	}
	v.actions.ErrLog(v.ConnectID(), v.conn.Close(), "close connect")
}

func (v *Connect) Run() {
	if !v.sync.On() {
		return
	}
	v.mux.RLock(func() {
		for _, fn := range v.openFunc {
			fn(v.ConnectID())
		}
	})
	internal.SetupPingPong(v.conn)
	go internal.PumpWrite(v, v.actions)
	internal.PumpRead(v, v.actions)
	v.mux.RLock(func() {
		for _, fn := range v.closeFunc {
			fn(v.ConnectID())
		}
	})
}
