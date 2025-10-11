/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/ws"
	"go.osspkg.com/goppy/v2/ws/event"
)

func main() {
	application := goppy.New("", "", "")
	application.Plugins(
		ws.WithClient(),
	)
	application.Plugins(
		plugins.Kind{
			Inject: NewController,
			Resolve: func(c *Controller, ctx xc.Context, cli ws.Client) error {
				cli.SetEventHandler(c.EventListener, 99, 1, 65000)
				cli.AddOnCloseFunc(func(cid string) {
					fmt.Println("server close connect")
					ctx.Close()
				})
				go c.Ticker(cli.BroadcastEvent)

				_, err := cli.Open("ws://127.0.0.1:10000/ws")
				return err
			},
		},
	)
	application.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Ticker(call func(event.Id, any) error) {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			call(1, []int{0})
		}
	}
}

func (v *Controller) EventListener(event event.Event, meta ws.Meta) error {
	var vv json.RawMessage
	if err := event.Decode(&vv); err != nil {
		return err
	}
	fmt.Println(">", "EventListener", meta.ConnectID(), event.ID(), string(vv))
	return nil
}
