/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/plugins/web"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/netutil/websocket"
)

func main() {
	application := goppy.New()
	application.Plugins(
		web.WithWebsocketClient(),
	)
	application.Plugins(
		plugins.Plugin{
			Inject: func() *Controller {
				return &Controller{}
			},
			Resolve: func(c *Controller, ctx app.Context, ws web.WebsocketClient) {
				wsc := ws.Create("ws://127.0.0.1:8088/ws")
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

func (v *Controller) Ticker(call func(id websocket.EventID, in interface{})) {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			call(1, []int{0})
		}
	}
}

func (v *Controller) EventListener(w websocket.CRequest, r websocket.CResponse, m websocket.CMeta) {
	var vv json.RawMessage
	if err := r.Decode(&vv); err != nil {
		fmt.Println(err)
	}
	fmt.Println("EventListener", m.ConnectID(), r.EventID(), string(vv))
}
