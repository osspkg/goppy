package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTPDebug(),
		http.WithHTTP(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes *http.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(http.ThrottlingMiddleware(1))
				router.Get("/user/{id}", c.User)
			},
		},
		plugins.Plugin{
			Inject: NewAdminController,
			Resolve: func(routes *http.RouterPool, c *AdminController) {
				router := routes.Get("admin")
				router.Get("/admin/{id}", c.Admin)

				apiColl := router.Collection("/api/v1", http.ThrottlingMiddleware(1))
				apiColl.Get("/data", c.Admin)
			},
		},
	)
	app.Run()

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()

	ctx.SetBody().Error(http.ErrMessage{
		HTTPCode:     400,
		InternalCode: "x1000",
		Message:      "Пользователь не найден",
		Ctx:          map[string]interface{}{"id": id},
	})

	ctx.Log().Infof("user - %d", id)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AdminController struct{}

func NewAdminController() *AdminController {
	return &AdminController{}
}

func (v *AdminController) Admin(ctx http.Ctx) {
	ctx.SetBody().Raw([]byte(ctx.URL().String()))
}
