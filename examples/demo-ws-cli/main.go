/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/web"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		web.WithWebsocketClient(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(c *Controller, ws web.WebsocketClient) error {
				wsc, err := ws.Create(context.TODO(), "ws://127.0.0.1:8088/ws")
				if err != nil {
					return err
				}

				wsc.Event(c.EventListener, 99)
				go c.Ticker(wsc.Encode)

				time.AfterFunc(30*time.Second, func() {
					wsc.Close()
				})

				go wsc.Run()

				return nil
			},
		},
	)
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Ticker(call func(id uint, in interface{})) {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			call(99, tt.Format(time.RFC3339))
		}
	}
}

func (v *Controller) EventListener(d web.WebsocketEventer, c web.WebsocketClientProcessor) error {
	fmt.Println("EventListener", c.ConnectID(), d.UniqueID(), d.EventID())
	return nil
}
