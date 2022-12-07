package main

import (
	"fmt"

	"github.com/dewep-online/goppy"
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
				router.Use(http.ThrottlingMiddleware(100))
				router.Get("/users", c.Users)

				api := router.Collection("/api/v1", http.ThrottlingMiddleware(100))
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
	ctx.SetBody(200).JSON(data)
}

func (v *Controller) User(ctx http.Ctx) {
	id, _ := ctx.Param("id").Int()

	err := v.db.Pool("main").Ping()
	if err != nil {
		ctx.Log().Errorf("db: %s", err.Error())
	}

	ctx.SetBody(400).ErrorJSON(fmt.Errorf("user not found"), "x1000", http.ErrCtx{"id": id})

	ctx.Log().Infof("user - %d", id)
}
