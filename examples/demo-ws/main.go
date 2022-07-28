package main

import (
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

				ws.Event(1, c.List)
				ws.Event(2, c.User)

				router.Get("/ws", ws.Handling)
			},
		},
	)
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) List(m http.Message, c http.Connection) error {
	list := make([]int, 0)
	if err := m.Decode(&list); err != nil {
		return err
	}
	list = append(list, 10, 19, 17, 15)
	return m.Encode(&list)
}

func (v *Controller) User(m http.Message, c http.Connection) error {
	id := c.UID()
	return m.Encode(&id)
}
