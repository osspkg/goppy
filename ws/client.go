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
	"time"

	"github.com/gorilla/websocket"
	"go.osspkg.com/do"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/ws/event"
)

type _client struct {
	events     *syncing.Map[event.Id, EventHandler]
	servers    *syncing.Map[string, *connect]
	ctx        context.Context
	cancel     context.CancelFunc
	openFuncs  *syncing.Slice[func(cid string)]
	closeFuncs *syncing.Slice[func(cid string)]
	wg         syncing.Group
}

func NewClient(ctx context.Context) Client {
	ctx, cancel := context.WithCancel(ctx)
	return &_client{
		events:     syncing.NewMap[event.Id, EventHandler](10),
		servers:    syncing.NewMap[string, *connect](10),
		ctx:        ctx,
		cancel:     cancel,
		openFuncs:  syncing.NewSlice[func(cid string)](2),
		closeFuncs: syncing.NewSlice[func(cid string)](2),
		wg:         syncing.NewGroup(ctx),
	}
}

func (v *_client) Up() error {
	return nil
}

func (v *_client) Down() error {
	v.CloseAll()
	return nil
}

func (v *_client) Open(url string, opts ...func(h http.Header, d *websocket.Dialer)) (string, error) {
	headers := make(http.Header)
	dial := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
	}

	for _, opt := range opts {
		opt(headers, dial)
	}

	conn, resp, err := dial.DialContext(v.ctx, url, headers)
	if err != nil {
		logx.Error("WS Client", "do", "open connect", "err", err, "url", url)
		return "", err
	}

	clientId := resp.Header.Get("Sec-WebSocket-Accept")

	v.wg.Background("open connect", func(ctx context.Context) {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logx.Error("WS Client", "do", "close connect body", "err", err, "url", url)
				return
			}
		}()

		c := newConnect(v.ctx, clientId, resp.Header, v, conn)

		c.AddOnCloseFunc(func(cid string) { v.delConn(cid) })
		c.AddOnOpenFunc(func(string) { v.addConn(c) })
		c.Run()
	})

	return clientId, nil
}

func (v *_client) GetEventHandler(eventId event.Id) (EventHandler, bool) {
	return v.events.Get(eventId)
}

func (v *_client) SetEventHandler(handler EventHandler, eventIDs ...event.Id) {
	for _, eventId := range eventIDs {
		v.events.Set(eventId, handler)
	}

}

func (v *_client) DelEventHandler(eventIDs ...event.Id) {
	for _, eventId := range eventIDs {
		v.events.Del(eventId)
	}
}

func (v *_client) CountConn() int {
	return v.servers.Size()
}

func (v *_client) addConn(conn *connect) {
	v.servers.Set(conn.ConnectID(), conn)

	for call := range v.openFuncs.Yield() {
		if err := do.Recovery(func() { call(conn.ConnectID()) }); err != nil {
			logx.Error("WS Server", "do", "run open func", "panic", err, "serverId", conn.ConnectID())
		}
	}
}

func (v *_client) delConn(serverId string) {
	v.servers.Del(serverId)

	go func() {
		for call := range v.closeFuncs.Yield() {
			if err := do.Recovery(func() { call(serverId) }); err != nil {
				logx.Error("WS Server", "do", "run close func", "panic", err, "serverId", serverId)
			}
		}
	}()
}

func (v *_client) AddOnCloseFunc(callback func(serverId string)) {
	v.closeFuncs.Append(callback)
}

func (v *_client) AddOnOpenFunc(callback func(serverId string)) {
	v.openFuncs.Append(callback)
}

func (v *_client) CloseConnect(serverId string) {
	conn, ok := v.servers.Extract(serverId)
	if !ok {
		return
	}

	conn.Close()
}

func (v *_client) CloseAll() {
	v.cancel()
	v.wg.Wait()
}

func (v *_client) BroadcastEvent(eventId event.Id, message any) error {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	ev.WithID(eventId)
	if err := ev.Encode(message); err != nil {
		logx.Error("WS Client", "do", "broadcast event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode message for id '%d': %w", eventId, err)
	}

	b, err := json.Marshal(ev)
	if err != nil {
		logx.Error("WS Client", "do", "broadcast event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode event for id '%d': %w", eventId, err)
	}

	go func() {
		for _, conn := range v.servers.Yield() {
			conn.SendRawMessage(b)
		}
	}()

	return nil
}

func (v *_client) SendEvent(eventId event.Id, message any, serverIDs ...string) error {
	ev := event.Pool.Get()
	defer event.Pool.Put(ev)

	ev.WithID(eventId)
	if err := ev.Encode(message); err != nil {
		logx.Error("WS Client", "do", "send event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode message for id '%d': %w", eventId, err)
	}

	b, err := json.Marshal(ev)
	if err != nil {
		logx.Error("WS Client", "do", "send event", "err", err, "eventId", eventId)
		return fmt.Errorf("encode event for id '%d': %w", eventId, err)
	}

	go func() {
		for _, clientId := range serverIDs {
			if conn, ok := v.servers.Get(clientId); ok {
				conn.SendRawMessage(b)
			}
		}
	}()

	return nil
}
