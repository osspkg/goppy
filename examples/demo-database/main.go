package main

import (
	"fmt"

	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/database"
	"github.com/osspkg/goppy/plugins/web"
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
	mdb database.MySQL
	sdb database.SQLite
}

func NewController(
	m database.MySQL,
	s database.SQLite,
) *Controller {
	return &Controller{
		mdb: m,
		sdb: s,
	}
}

func (v *Controller) Users(ctx web.Context) {
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int()

	err := v.mdb.Pool("main").Ping()
	if err != nil {
		ctx.ErrorJSON(500, err, web.ErrCtx{"id": id})
		return
	}

	err = v.sdb.Pool("main").Ping()
	if err != nil {
		ctx.ErrorJSON(500, err, web.ErrCtx{"id": id})
		return
	}

	ctx.ErrorJSON(400, fmt.Errorf("user not found"), web.ErrCtx{"id": id})

	ctx.Log().Infof("user - %d", id)
}
