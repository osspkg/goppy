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
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes http.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(middlewares.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", middlewares.ThrottlingMiddleware(100))
				api.Get("/user/{id}", c.User)
			},
		},
	)
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx http.Ctx) {
	data := []int64{1, 2, 3, 4}
	ctx.SetBody(200).JSON(data)
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()
	ctx.SetBody(200).String("user id: %d", id)
	ctx.Log().Infof("user - %d", id)
}
