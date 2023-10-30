/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package websocket

import (
	"context"
	"encoding/json"
	"net/http"

	ws "github.com/gorilla/websocket"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
)

type (
	cli struct {
		url       string
		id        string
		header    http.Header
		events    map[EventID]ClientHandler
		conn      *ws.Conn
		logger    log.Logger
		busBuf    chan []byte
		ctx       context.Context
		cancel    context.CancelFunc
		openFunc  []func(cid string)
		closeFunc []func(cid string)
		sync      iosync.Switch
		mux       iosync.Lock
	}

	Client interface {
		SendEvent(eid EventID, in interface{})
		ConnectID() string
		Header(key, value string)
		SetHandler(call ClientHandler, eids ...EventID)
		DelHandler(eids ...EventID)
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
		Close()
		DialAndListen() error
	}

	ClientOption interface {
		Header(key, value string)
	}
)

func NewClient(ctx context.Context, url string, l log.Logger, opts ...func(ClientOption)) Client {
	c, cancel := context.WithCancel(ctx)
	wcli := &cli{
		url:       url,
		id:        "",
		header:    make(http.Header),
		events:    make(map[EventID]ClientHandler, 10),
		conn:      nil,
		logger:    l,
		busBuf:    make(chan []byte, busBufferSize),
		ctx:       c,
		cancel:    cancel,
		openFunc:  make([]func(string), 0, 2),
		closeFunc: make([]func(string), 0, 2),
		sync:      iosync.NewSwitch(),
		mux:       iosync.NewLock(),
	}
	for _, opt := range opts {
		opt(wcli)
	}
	return wcli
}

func (v *cli) ErrLog(cid string, err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}
	v.logger.WithFields(log.Fields{"cid": cid, "err": err.Error()}).Errorf(msg, args...)
}

func (v *cli) ErrLogMessage(cid string, msg string, args ...interface{}) {
	v.logger.WithFields(log.Fields{"cid": cid}).Errorf(msg, args...)
}

func (v *cli) SetHandler(call ClientHandler, eids ...EventID) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			v.events[eid] = call
		}
	})
}

func (v *cli) DelHandler(eids ...EventID) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			delete(v.events, eid)
		}
	})
}

func (v *cli) GetHandler(eid EventID) (h ClientHandler, ok bool) {
	v.mux.RLock(func() {
		h, ok = v.events[eid]
	})
	return
}

func (v *cli) Header(key, value string) {
	v.mux.Lock(func() {
		v.header.Set(key, value)
	})
}

func (v *cli) ConnectID() string {
	return v.id
}

func (v *cli) connect() *ws.Conn {
	return v.conn
}

func (v *cli) cancelFunc() context.CancelFunc {
	return v.cancel
}

func (v *cli) done() <-chan struct{} {
	return v.ctx.Done()
}

func (v *cli) readBus() <-chan []byte {
	return v.busBuf
}

func (v *cli) WriteToBus(b []byte) {
	if v.sync.IsOff() {
		return
	}
	if len(b) == 0 {
		return
	}
	select {
	case v.busBuf <- b:
	default:
		v.ErrLogMessage(v.id, "write chan is full")
	}
}

func (v *cli) SendEvent(eid EventID, in interface{}) {
	getEventModel(func(ev *event) {
		ev.ID = eid
		ev.Encode(in)
		b, err := json.Marshal(ev)
		if err != nil {
			v.ErrLog(v.ConnectID(), err, "[ws] encode message: %d", eid)
			return
		}
		v.WriteToBus(b)
	})
}

func (v *cli) callHandler(b []byte) {
	getEventModel(func(ev *event) {
		if err := json.Unmarshal(b, ev); err != nil {
			v.ErrLog(v.ConnectID(), err, "[ws] decode message")
			return
		}
		call, ok := v.GetHandler(ev.EventID())
		if !ok {
			return
		}
		call(ev, ev, v)
	})
}

func (v *cli) OnClose(cb func(cid string)) {
	v.mux.Lock(func() {
		v.closeFunc = append(v.closeFunc, cb)
	})
}

func (v *cli) OnOpen(cb func(cid string)) {
	v.mux.Lock(func() {
		v.openFunc = append(v.openFunc, cb)
	})
}

func (v *cli) Close() {
	if !v.sync.Off() {
		return
	}
	v.cancel()
}

func (v *cli) DialAndListen() error {
	if !v.sync.On() {
		return errOneOpenConnect
	}
	defer v.sync.Off()

	var (
		err  error
		resp *http.Response
	)

	if v.conn, resp, err = ws.DefaultDialer.DialContext(v.ctx, v.url, v.header); err != nil {
		v.ErrLog(v.ConnectID(), err, "open connect [%s]", v.url)
		return err
	} else {
		v.id = resp.Header.Get("Sec-WebSocket-Accept")
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			v.ErrLog(v.ConnectID(), err, "close body connect [%s]", v.url)
		}
	}()

	v.mux.RLock(func() {
		for _, fn := range v.openFunc {
			fn(v.ConnectID())
		}
	})
	setupPingPong(v.connect())
	go pumpWrite(v, v)
	pumpRead(v, v)
	v.mux.RLock(func() {
		for _, fn := range v.closeFunc {
			fn(v.ConnectID())
		}
	})
	return nil
}
