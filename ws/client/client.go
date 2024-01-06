/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/ws/internal"
	"go.osspkg.com/goppy/xlog"
)

type (
	cli struct {
		url       string
		id        string
		header    http.Header
		events    map[event.Id]Handler
		conn      *websocket.Conn
		logger    xlog.Logger
		busBuf    chan []byte
		ctx       context.Context
		cancel    context.CancelFunc
		openFunc  []func(cid string)
		closeFunc []func(cid string)
		sync      iosync.Switch
		mux       iosync.Lock
	}

	Client interface {
		SendEvent(eid event.Id, in interface{})
		ConnectID() string
		Header(key, value string)
		SetHandler(call Handler, eids ...event.Id)
		DelHandler(eids ...event.Id)
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
		Close()
		DialAndListen() error
	}

	Option interface {
		Header(key, value string)
	}
)

func New(ctx context.Context, url string, l xlog.Logger, opts ...func(Option)) Client {
	c, cancel := context.WithCancel(ctx)
	wcli := &cli{
		url:       url,
		id:        "",
		header:    make(http.Header),
		events:    make(map[event.Id]Handler, 10),
		conn:      nil,
		logger:    l,
		busBuf:    make(chan []byte, internal.BusBufferSize),
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
	v.logger.WithFields(xlog.Fields{"cid": cid, "err": err.Error()}).Errorf(msg, args...)
}

func (v *cli) ErrLogMessage(cid string, msg string, args ...interface{}) {
	v.logger.WithFields(xlog.Fields{"cid": cid}).Errorf(msg, args...)
}

func (v *cli) SetHandler(call Handler, eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			v.events[eid] = call
		}
	})
}

func (v *cli) DelHandler(eids ...event.Id) {
	v.mux.Lock(func() {
		for _, eid := range eids {
			delete(v.events, eid)
		}
	})
}

func (v *cli) GetHandler(eid event.Id) (h Handler, ok bool) {
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

func (v *cli) Connect() *websocket.Conn {
	return v.conn
}

func (v *cli) CancelFunc() context.CancelFunc {
	return v.cancel
}

func (v *cli) Done() <-chan struct{} {
	return v.ctx.Done()
}

func (v *cli) ReadBus() <-chan []byte {
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

func (v *cli) SendEvent(eid event.Id, in interface{}) {
	event.GetMessage(func(ev *event.Message) {
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

func (v *cli) CallHandler(b []byte) {
	event.GetMessage(func(ev *event.Message) {
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
		return internal.ErrOneOpenConnect
	}
	defer v.sync.Off()

	var (
		err  error
		resp *http.Response
	)

	if v.conn, resp, err = websocket.DefaultDialer.DialContext(v.ctx, v.url, v.header); err != nil {
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
	internal.SetupPingPong(v.Connect())
	go internal.PumpWrite(v, v)
	internal.PumpRead(v, v)
	v.mux.RLock(func() {
		for _, fn := range v.closeFunc {
			fn(v.ConnectID())
		}
	})
	return nil
}
