package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTP(),
		http.WithWebsocketServer(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes http.RouterPool, c *Controller, ws http.WebsocketServer) {
				router := routes.Main()
				router.Use(http.ThrottlingMiddleware(100))

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
	list map[string]http.WebsocketServerProcessor
	mux  sync.RWMutex
}

func NewController() *Controller {
	c := &Controller{
		list: make(map[string]http.WebsocketServerProcessor),
	}
	go c.Timer()
	return c
}

func (v *Controller) Event99(ev http.WebsocketEventer, c http.WebsocketServerProcessor) error {
	var data string
	if err := ev.Decode(&data); err != nil {
		return err
	}
	c.EncodeEvent(ev, &data)
	fmt.Println(c.CID(), "Event99", ev.EventID(), ev.UniqueID())
	return nil
}

func (v *Controller) OneEvent(ev http.WebsocketEventer, c http.WebsocketServerProcessor) error {
	list := make([]int, 0)
	if err := ev.Decode(&list); err != nil {
		return err
	}
	list = append(list, 10, 19, 17, 15)
	c.EncodeEvent(ev, &list)
	fmt.Println(c.CID(), "OneEvent", ev.EventID(), ev.UniqueID())
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

func (v *Controller) MultiEvent(d http.WebsocketEventer, c http.WebsocketServerProcessor) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	switch d.EventID() {
	case 11:
		v.list[c.CID()] = c
		fmt.Println("add", c.CID())
		c.OnClose(func(cid string) {
			fmt.Println("close", cid)
			delete(v.list, cid)
		})
	case 13:
		fmt.Println("del", c.CID())
		delete(v.list, c.CID())
	}
	return nil
}
