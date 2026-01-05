/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/do"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/web"
	"go.osspkg.com/goppy/v3/ws/event"
	"go.osspkg.com/goppy/v3/ws/internal"
)

type _server struct {
	cancel     context.CancelFunc
	upgrade    *websocket.Upgrader
	clients    *syncing.Map[string, *connect]
	events     *syncing.Map[event.Id, EventHandler]
	guard      *syncing.Slice[Guard]
	openFuncs  *syncing.Slice[func(cid string)]
	closeFuncs *syncing.Slice[func(cid string)]
	wg         syncing.Group
}

func NewServer(ctx context.Context, opts ...func(u *websocket.Upgrader)) Server {
	upgrade := internal.NewUpgrade()
	ctx, cancel := context.WithCancel(ctx)
	for _, opt := range opts {
		opt(upgrade)
	}

	return &_server{
		upgrade:    upgrade,
		cancel:     cancel,
		clients:    syncing.NewMap[string, *connect](10),
		events:     syncing.NewMap[event.Id, EventHandler](10),
		guard:      syncing.NewSlice[Guard](2),
		openFuncs:  syncing.NewSlice[func(cid string)](2),
		closeFuncs: syncing.NewSlice[func(cid string)](2),
		wg:         syncing.NewGroup(ctx),
	}
}

func (v *_server) Up() error {
	return nil
}

func (v *_server) Down() error {
	v.CloseAll()
	return nil
}

func (v *_server) CountConn() int {
	return v.clients.Size()
}

func (v *_server) GetEventHandler(eventId event.Id) (h EventHandler, ok bool) {
	return v.events.Get(eventId)
}

func (v *_server) SetEventHandler(handler EventHandler, eventIDs ...event.Id) {
	for _, eventId := range eventIDs {
		v.events.Set(eventId, handler)
	}
}

func (v *_server) DelEventHandler(eventIDs ...event.Id) {
	for _, eventId := range eventIDs {
		v.events.Del(eventId)
	}
}

func (v *_server) AddGuard(g Guard) {
	v.guard.Append(g)
}

func (v *_server) callGuards(clientId string, head http.Header) error {
	for guardFunc := range v.guard.Yield() {
		if guardFunc == nil {
			continue
		}
		if err := guardFunc(clientId, head); err != nil {
			return err
		}
	}

	return nil
}

func (v *_server) BroadcastEvent(eventId event.Id, message any) error {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	ev.WithID(eventId)
	if err := ev.Encode(message); err != nil {
		logx.Error("WS Server", "do", "broadcast event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode message for id '%d': %w", eventId, err)
	}

	b, err := json.Marshal(ev)
	if err != nil {
		logx.Error("WS Server", "do", "broadcast event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode event for id '%d': %w", eventId, err)
	}

	go func() {
		for _, conn := range v.clients.Yield() {
			conn.SendRawMessage(b)
		}
	}()

	return nil
}

func (v *_server) SendEvent(eventId event.Id, message any, clientIDs ...string) error {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	ev.WithID(eventId)
	if err := ev.Encode(message); err != nil {
		logx.Error("WS Server", "do", "send event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode message for id '%d': %w", eventId, err)
	}

	b, err := json.Marshal(ev)
	if err != nil {
		logx.Error("WS Server", "do", "send event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode event for id '%d': %w", eventId, err)
	}

	go func() {
		for _, clientId := range clientIDs {
			if conn, ok := v.clients.Get(clientId); ok {
				conn.SendRawMessage(b)
			}
		}
	}()

	return nil
}

func (v *_server) addConn(conn *connect) {
	v.clients.Set(conn.ConnectID(), conn)

	go func() {
		for call := range v.openFuncs.Yield() {
			if err := do.Recovery(func() { call(conn.ConnectID()) }); err != nil {
				logx.Error("WS Server", "do", "run open func", "panic", err, "clientId", conn.ConnectID())
			}
		}
	}()
}

func (v *_server) delConn(clientId string) {
	v.clients.Del(clientId)

	go func() {
		for call := range v.closeFuncs.Yield() {
			if err := do.Recovery(func() { call(clientId) }); err != nil {
				logx.Error("WS Server", "do", "run close func", "panic", err, "clientId", clientId)
			}
		}
	}()
}

func (v *_server) AddOnCloseFunc(call func(clientId string)) {
	v.closeFuncs.Append(call)
}

func (v *_server) AddOnOpenFunc(call func(clientId string)) {
	v.openFuncs.Append(call)
}

func (v *_server) CloseOne(clientId string) {
	conn, ok := v.clients.Extract(clientId)
	if !ok {
		return
	}

	conn.Close()
}

func (v *_server) CloseAll() {
	v.cancel()
	v.wg.Wait()
}

func (v *_server) Handling(ctx web.Ctx) {
	v.HandlingHTTP(ctx.Response(), ctx.Request())
}

func (v *_server) HandlingHTTP(w http.ResponseWriter, r *http.Request) {
	v.wg.Run("handling", func(ctx context.Context) {
		clientId := r.Header.Get("Sec-Websocket-Key")

		defer func() {
			if err := r.Body.Close(); err != nil {
				logx.Error("WS Server", "do", "close connect body", "err", err, "clientId", clientId)
			}
		}()

		upgrade, err := v.upgrade.Upgrade(w, r, nil)
		if err != nil {
			logx.Error("WS Server", "do", "handling: upgrade new connect", "err", err, "clientId", clientId)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = v.callGuards(clientId, r.Header); err != nil {
			logx.Error("WS Server", "do", "handling: guard", "err", err, "clientId", clientId)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		ctx, cancel := xc.Join(ctx, r.Context())
		defer cancel()

		conn := newConnect(ctx, clientId, r.Header, v, upgrade)
		conn.AddOnOpenFunc(func(string) { v.addConn(conn) })
		conn.AddOnCloseFunc(func(cid string) { v.delConn(cid) })
		conn.Run()
	})
}
