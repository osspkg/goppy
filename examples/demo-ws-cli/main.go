package main

import (
	"fmt"
	"time"

	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithWebsocketClient(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(c *Controller, ws http.WebsocketClient) error {
				wsc, err := ws.Create("ws://127.0.0.1:8088/ws")
				if err != nil {
					return err
				}
				defer wsc.Close()

				wsc.Event(c.EventListener, 99)
				go c.Ticker(wsc.Encode)

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

func (v *Controller) EventListener(d http.WebsocketEventer, c http.WebsocketClientProcessor) error {
	fmt.Println("EventListener", c.CID(), d.UniqueID(), d.EventID())
	return nil
}
