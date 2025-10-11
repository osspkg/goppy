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
	"go.osspkg.com/goppy/v2/ws/internal"
)

type (
	eventResolver interface {
		GetEventHandler(eid event.Id) (EventHandler, bool)
	}

	connect struct {
		id         string
		header     http.Header
		resolver   eventResolver
		conn       *websocket.Conn
		dataC      chan []byte
		ctx        context.Context
		cancel     context.CancelFunc
		openFuncs  *syncing.Slice[func(cid string)]
		closeFuncs *syncing.Slice[func(cid string)]
		status     syncing.Switch
	}
)

func newConnect(ctx context.Context, id string, head http.Header, r eventResolver, conn *websocket.Conn) *connect {
	ctx, cancel := context.WithCancel(ctx)
	return &connect{
		id:         id,
		header:     head,
		resolver:   r,
		conn:       conn,
		dataC:      make(chan []byte, internal.BusBufferSize),
		ctx:        ctx,
		cancel:     cancel,
		closeFuncs: syncing.NewSlice[func(cid string)](2),
		openFuncs:  syncing.NewSlice[func(cid string)](2),
		status:     syncing.NewSwitch(),
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

	for call := range v.openFuncs.Yield() {
		if err := do.Recovery(func() { call(v.ConnectID()) }); err != nil {
			logx.Error("WS Connect", "do", "run open func", "panic", err, "cid", v.ConnectID())
		}
	}

	internal.SetupPingPong(v.conn)

	wg := syncing.NewGroup(v.ctx)
	wg.Background("pump write", func(_ context.Context) { internal.PumpWrite(v) })
	wg.Background("pump read", func(_ context.Context) { internal.PumpRead(v) })
	wg.Wait()

	for call := range v.closeFuncs.Yield() {
		if err := do.Recovery(func() { call(v.ConnectID()) }); err != nil {
			logx.Error("WS Connect", "do", "run close func", "panic", err, "cid", v.ConnectID())
		}
	}
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

func (v *connect) AddOnCloseFunc(call func(cid string)) {
	v.closeFuncs.Append(call)
}

func (v *connect) AddOnOpenFunc(call func(cid string)) {
	v.openFuncs.Append(call)
}

func (v *connect) SendMessageChan() <-chan []byte {
	return v.dataC
}

func (v *connect) ReceiveMessage(b []byte) {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	if err := json.Unmarshal(b, ev); err != nil {
		logx.Error("WS Connect", "do", "decode receive message", "err", err, "cid", v.ConnectID())
		return
	}

	call, ok := v.resolver.GetEventHandler(ev.ID())
	if !ok {
		ev.WithError(internal.ErrUnknownEventID)
	} else if err := call(ev, v); err != nil {
		ev.WithError(err)
	}

	if bb, err := json.Marshal(ev); err != nil {
		logx.Error("WS Connect", "do", "encode receive message", "err", err, "cid", v.ConnectID())
	} else {
		v.SendRawMessage(bb)
	}
}

func (v *connect) SendRawMessage(message []byte) {
	if len(message) == 0 {
		return
	}
	select {
	case v.dataC <- message:
	default:
		logx.Error("WS Connect", "do", "send message", "err", "write chan is full", "cid", v.ConnectID())
	}
	return
}

func (v *connect) SendEvent(eventId event.Id, message any) {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	ev.WithID(eventId)
	if err := ev.Encode(message); err != nil {
		logx.Error("WS Connect", "do", "encode message", "err", err, "cid", v.ConnectID())
		return
	}

	b, err := json.Marshal(ev)
	if err != nil {
		logx.Error("WS Connect", "do", "encode event", "err", err, "cid", v.ConnectID())
		return
	}

	v.SendRawMessage(b)
}
