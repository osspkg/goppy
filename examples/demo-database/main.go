package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/middlewares"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/database"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTPDebug(),
		http.WithHTTP(),
		database.WithMySQL(),
		database.WithSQLite(),
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

type Controller struct {
	db database.MySQL
}

func NewController(v database.MySQL) *Controller {
	return &Controller{
		db: v,
	}
}

func (v *Controller) Users(ctx http.Ctx) {
	data := []int64{1, 2, 3, 4}
	ctx.SetBody().JSON(data)
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()

	err := v.db.Pool("main").Ping()
	if err != nil {
		ctx.Log().Errorf("db: %s", err.Error())
	}

	ctx.SetBody().Error(http.ErrMessage{
		HTTPCode:     400,
		InternalCode: "x1000",
		Message:      "user not found",
		Ctx:          map[string]interface{}{"id": id},
	})

	ctx.Log().Infof("user - %d", id)
}
