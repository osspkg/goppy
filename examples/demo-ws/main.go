package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/middlewares"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTP(),
		http.WithWebsocket(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes http.RouterPool, c *Controller, ws http.WebSocket) {
				router := routes.Main()
				router.Use(middlewares.ThrottlingMiddleware(100))

				ws.Event(c.OneEvent, 1, 2)
				ws.Event(c.MultiEvent, 11, 13)

				router.Get("/ws", ws.Handling)
			},
		},
	)
	app.Run()
}

type Controller struct {
	list map[string]http.Processor
	mux  sync.RWMutex
}

func NewController() *Controller {
	c := &Controller{
		list: make(map[string]http.Processor),
	}
	go c.Timer()
	return c
}

func (v *Controller) OneEvent(d http.Eventer, c http.Processor) error {
	list := make([]int, 0)
	if err := d.Decode(&list); err != nil {
		return err
	}
	list = append(list, 10, 19, 17, 15)
	c.Encode(1, &list)
	return nil
}

func (v *Controller) Timer() {
	t := time.NewTicker(time.Second * 3)
	defer t.Stop()

	for {
		select {
		case tt := <-t.C:
			v.mux.RLock()
			for _, p := range v.list {
				p.Encode(12, tt.Format(time.RFC3339))
			}
			v.mux.RUnlock()
		}
	}
}

func (v *Controller) MultiEvent(d http.Eventer, c http.Processor) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	switch d.EventID() {
	case 11:
		v.list[c.CID()] = c
		fmt.Println("add", c.CID())
		c.OnClose(func(cid string) {
			fmt.Println("del", cid)
			delete(v.list, cid)
		})
	case 13:
		fmt.Println("del", c.CID())
		delete(v.list, c.CID())
	}
	return nil
}
