/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/web"
	"go.osspkg.com/goppy/v3/ws"
	"go.osspkg.com/goppy/v3/ws/event"
)

func main() {
	app := goppy.New("goppy_ws_server", "v1.0.0", "")
	app.Plugins(
		web.WithServer(),
		ws.WithServer(),
	)
	app.Plugins(
		func(wss ws.Server) *Controller {
			return NewController(wss)
		},
		func(routes web.ServerPool, c *Controller, wss ws.Server) {
			router, ok := routes.Main()
			if !ok {
				return
			}

			wss.SetEventHandler(c.Event1, 1)
			wss.SetEventHandler(c.MultiEvent, 2, 3)

			router.Get("/ws", func(ctx web.Ctx) {
				wss.Handling(ctx)
			})

			router.Get("/", func(ctx web.Ctx) {
				b, err := os.ReadFile("./index.html")
				if err != nil {
					ctx.Error(http.StatusInternalServerError, err)
					return
				}
				ctx.Bytes(http.StatusOK, b)
			})
		},
	)
	app.Run()
}

type (
	pipe interface {
		BroadcastEvent(eid event.Id, m any) (err error)
		SendEvent(eid event.Id, m any, cids ...string) (err error)
		AddOnCloseFunc(cb func(cid string))
		AddOnOpenFunc(cb func(cid string))
	}
	Controller struct {
		pipe pipe
		list *syncing.Map[string, struct{}]
	}
)

func NewController(s pipe) *Controller {
	ctrl := &Controller{
		list: syncing.NewMap[string, struct{}](2),
		pipe: s,
	}

	ctrl.pipe.AddOnOpenFunc(func(cid string) {
		ctrl.list.Set(cid, struct{}{})
		fmt.Println(">", cid, "MultiEvent Add")
	})

	ctrl.pipe.AddOnCloseFunc(func(cid string) {
		ctrl.list.Del(cid)
		fmt.Println(">", cid, "MultiEvent Close")
	})

	go ctrl.Timer()

	return ctrl
}

func (v *Controller) Event1(event event.Event, meta ws.Meta) error {
	list := make([]int, 0)
	if err := event.Decode(&list); err != nil {
		return err
	}

	list = []int{10, 19, 17, 15}

	if err := event.Encode(&list); err != nil {
		fmt.Println(">", meta.ConnectID(), "Event1", event.ID(), "err", err.Error())
		return err
	}

	fmt.Println(">", meta.ConnectID(), "Event1", event.ID())

	return nil
}

func (v *Controller) Timer() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			if err := v.pipe.BroadcastEvent(99, tt.Format(time.RFC3339)); err != nil {
				fmt.Println(">", "Timer", "err", err.Error())
			}
		}
	}
}

func (v *Controller) MultiEvent(event event.Event, meta ws.Meta) error {
	fmt.Println(">", meta.ConnectID(), "MultiEvent", event.ID())

	switch event.ID() {
	case 2:
		return event.Encode("event 2")

	default:
		return fmt.Errorf("event %d not found", event.ID())
	}
}
