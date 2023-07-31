/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/web"
)

func main() {
	app := goppy.New()
	app.Plugins(
		web.WithHTTP(),
		web.WithWebsocketServer(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes web.RouterPool, c *Controller, ws web.WebsocketServer) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))

				ws.Event(c.Event99, 99)
				ws.Event(c.OneEvent, 1, 2)
				ws.Event(c.MultiEvent, 11, 13)

				router.Get("/ws", ws.Handling)
			},
		},
	)
	app.Run()
}

type Controller struct {
	list map[string]web.WebsocketServerProcessor
	mux  sync.RWMutex
}

func NewController() *Controller {
	c := &Controller{
		list: make(map[string]web.WebsocketServerProcessor),
	}
	go c.Timer()
	return c
}

func (v *Controller) Event99(ev web.WebsocketEventer, c web.WebsocketServerProcessor) error {
	var data string
	if err := ev.Decode(&data); err != nil {
		return err
	}
	c.EncodeEvent(ev, &data)
	fmt.Println(c.ConnectID(), "Event99", ev.EventID(), ev.UniqueID())
	return nil
}

func (v *Controller) OneEvent(ev web.WebsocketEventer, c web.WebsocketServerProcessor) error {
	list := make([]int, 0)
	if err := ev.Decode(&list); err != nil {
		return err
	}
	list = append(list, 10, 19, 17, 15)
	c.EncodeEvent(ev, &list)
	fmt.Println(c.ConnectID(), "OneEvent", ev.EventID(), ev.UniqueID())
	return nil
}

func (v *Controller) Timer() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			v.muxRLock(func() {
				for _, p := range v.list {
					p.Encode(12, tt.Format(time.RFC3339))
					fmt.Println("Timer", p.ConnectID())
				}
			})
		}
	}
}

func (v *Controller) MultiEvent(d web.WebsocketEventer, c web.WebsocketServerProcessor) error {
	switch d.EventID() {
	case 11:
		v.muxLock(func() {
			v.list[c.ConnectID()] = c
			fmt.Println("MultiEvent Add", c.ConnectID())
		})

		c.OnClose(func(cid string) {
			v.muxLock(func() {
				delete(v.list, cid)
				fmt.Println("MultiEvent Close", cid)
			})
		})

	case 13:
		v.muxLock(func() {
			delete(v.list, c.ConnectID())
			fmt.Println("MultiEvent Del", c.ConnectID())
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
