package main

import (
	"fmt"

	"github.com/deweppro/goppy"
	"github.com/deweppro/goppy/plugins"
	"github.com/deweppro/goppy/plugins/database"
	"github.com/deweppro/goppy/plugins/web"
)

func main() {

	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		web.WithHTTPDebug(),
		web.WithHTTP(),
		database.WithMySQL(),
		database.WithSQLite(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes web.RouterPool, c *Controller) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", web.ThrottlingMiddleware(100))
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

func (v *Controller) Users(ctx web.Context) {
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int()

	err := v.db.Pool("main").Ping()
	if err != nil {
		ctx.Log().Errorf("db: %s", err.Error())
	}

	ctx.ErrorJSON(400, fmt.Errorf("user not found"), web.ErrCtx{"id": id})

	ctx.Log().Infof("user - %d", id)
}
