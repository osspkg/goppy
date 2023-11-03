/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"sync"
	"time"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/web"
	"go.osspkg.com/goppy/ws"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/ws/server"
)

func main() {
	app := goppy.New()
	app.Plugins(
		web.WithHTTP(),
		ws.WithWebsocketServer(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: func(ws ws.WebsocketServer) *Controller {
				return NewController(ws)
			},
			Resolve: func(routes web.RouterPool, c *Controller, wss ws.WebsocketServer) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))

				wss.SetHandler(c.Event99, 99)
				wss.SetHandler(c.OneEvent, 1, 2)
				wss.SetHandler(c.MultiEvent, 11, 13)

				router.Get("/ws", func(ctx web.Context) {
					wss.Handling(ctx.Response(), ctx.Request())
				})
			},
		},
	)
	app.Run()
}

type (
	sender interface {
		SendEvent(eid event.Id, m interface{}, cids ...string)
		Broadcast(eid event.Id, m interface{})
	}
	Controller struct {
		list   map[string]struct{}
		sender sender
		mux    sync.RWMutex
	}
)

func NewController(s sender) *Controller {
	c := &Controller{
		list:   make(map[string]struct{}),
		sender: s,
	}
	go c.Timer()
	return c
}

func (v *Controller) Event99(w server.Response, r server.Request, m server.Meta) error {
	var data string
	if err := r.Decode(&data); err != nil {
		return err
	}
	w.Encode(&data)
	fmt.Println(m.ConnectID(), "Event99", r.EventID())
	return nil
}

func (v *Controller) OneEvent(w server.Response, r server.Request, m server.Meta) error {
	list := make([]int, 0)
	if err := r.Decode(&list); err != nil {
		return err
	}
	list = append(list, 10, 19, 17, 15)
	w.Encode(&list)
	fmt.Println(m.ConnectID(), "OneEvent", r.EventID())
	return nil
}

func (v *Controller) Timer() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			v.muxRLock(func() {
				for cid := range v.list {
					v.sender.SendEvent(12, tt.Format(time.RFC3339), cid)
					fmt.Println("Timer", cid)
				}
			})
			v.sender.Broadcast(99, tt.Unix())
		}
	}
}

func (v *Controller) MultiEvent(w server.Response, r server.Request, m server.Meta) error {
	switch r.EventID() {
	case 11:
		v.muxLock(func() {
			v.list[m.ConnectID()] = struct{}{}
			fmt.Println("MultiEvent Add", m.ConnectID())
		})

		m.OnClose(func(cid string) {
			v.muxLock(func() {
				delete(v.list, cid)
				fmt.Println("MultiEvent Close", cid)
			})
		})

	case 13:
		v.muxLock(func() {
			delete(v.list, m.ConnectID())
			fmt.Println("MultiEvent Del", m.ConnectID())
		})

	}
	return nil
}

func (v *Controller) muxLock(cb func()) {
	v.mux.Lock()
	cb()
	v.mux.Unlock()
}

func (v *Controller) muxRLock(cb func()) {
	v.mux.RLock()
	cb()
	v.mux.RUnlock()
}
