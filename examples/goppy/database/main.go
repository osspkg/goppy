/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/metrics"
	"go.osspkg.com/goppy/ormmysql"
	"go.osspkg.com/goppy/ormsqlite"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/web"
)

func main() {

	app := goppy.New()
	app.Plugins(
		metrics.WithMetrics(),
		web.WithHTTP(),
		ormmysql.WithMySQL(),
		ormsqlite.WithSQLite(),
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
	mdb ormmysql.MySQL
	sdb ormsqlite.SQLite
}

func NewController(
	m ormmysql.MySQL,
	s ormsqlite.SQLite,
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
	id, _ := ctx.Param("id").Int() //nolint: errcheck

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
