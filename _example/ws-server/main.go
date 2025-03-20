/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"time"

	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/goppy/v2/ws"
	"go.osspkg.com/goppy/v2/ws/event"
)

func main() {
	app := goppy.New("goppy_ws_server", "v1.0.0", "")
	app.Plugins(
		web.WithServer(),
		ws.WithServer(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: func(ws ws.Server) *Controller {
				return NewController(ws)
			},
			Resolve: func(routes web.RouterPool, c *Controller, wss ws.Server) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))

				wss.SetEventHandler(c.Event1, 1)
				wss.SetEventHandler(c.MultiEvent, 2, 3)

				router.Get("/ws", func(ctx web.Context) {
					wss.HandlingHTTP(ctx.Response(), ctx.Request())
				})
			},
		},
	)
	app.Run()
}

type (
	sender interface {
		BroadcastEvent(eid event.Id, m any) (err error)
		SendEvent(eid event.Id, m any, cids ...string) (err error)
		OnClose(cb func(cid string))
		OnOpen(cb func(cid string))
	}
	Controller struct {
		list   map[string]struct{}
		sender sender
		mux    syncing.Lock
	}
)

func NewController(s sender) *Controller {
	c := &Controller{
		list:   make(map[string]struct{}),
		sender: s,
		mux:    syncing.NewLock(),
	}
	c.sender.OnOpen(func(cid string) {
		c.mux.Lock(func() {
			c.list[cid] = struct{}{}
			fmt.Println(">", cid, "MultiEvent Add")
		})
	})
	c.sender.OnClose(func(cid string) {
		c.mux.Lock(func() {
			delete(c.list, cid)
			fmt.Println(">", cid, "MultiEvent Close")
		})
	})
	go c.Timer()
	return c
}

func (v *Controller) Event1(event event.Event, meta ws.Meta) error {
	list := make([]int, 0)
	if err := event.Decode(&list); err != nil {
		return err
	}
	list = []int{10, 19, 17, 15}
	event.Encode(&list)
	fmt.Println(">", meta.ConnectID(), "Event1", event.ID())
	return nil
}

func (v *Controller) Timer() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			v.sender.BroadcastEvent(99, tt.Format(time.RFC3339))
		}
	}
}

func (v *Controller) MultiEvent(event event.Event, meta ws.Meta) error {
	switch event.ID() {
	case 2:
		event.Encode("event 2")

	case 3:
		event.WithError(fmt.Errorf("event 3"))
	}
	fmt.Println(">", meta.ConnectID(), "MultiEvent", event.ID())
	return nil
}
