/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/ws"
	"go.osspkg.com/goppy/ws/client"
	"go.osspkg.com/goppy/ws/event"
	"go.osspkg.com/goppy/xc"
)

func main() {
	application := goppy.New()
	application.Plugins(
		ws.WithWebsocketClient(),
	)
	application.Plugins(
		plugins.Plugin{
			Inject: func() *Controller {
				return &Controller{}
			},
			Resolve: func(c *Controller, ctx xc.Context, wws ws.WebsocketClient) {
				wsc := wws.Create("ws://127.0.0.1:8088/ws")
				wsc.SetHandler(c.EventListener, 99, 1, 65000)
				go c.Ticker(wsc.SendEvent)
				wsc.OnClose(func(cid string) {
					fmt.Println("server close connect")
					ctx.Close()
				})
			},
		},
	)
	application.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Ticker(call func(id event.Id, in interface{})) {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			call(1, []int{0})
		}
	}
}

func (v *Controller) EventListener(w client.Request, r client.Response, m client.Meta) {
	var vv json.RawMessage
	if err := r.Decode(&vv); err != nil {
		fmt.Println(err)
	}
	fmt.Println("EventListener", m.ConnectID(), r.EventID(), string(vv))
}
