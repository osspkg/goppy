/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/orm"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/logx"
)

func main() {

	app := goppy.New("goppy_database", "v1.0.0", "")
	app.Plugins(
		web.WithServer(),
		orm.WithMysql(),
		orm.WithSqlite(),
		orm.WithPGSql(),
		orm.WithORM(),
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
	orm orm.ORM
}

func NewController(orm orm.ORM) *Controller {
	return &Controller{
		orm: orm,
	}
}

func (v *Controller) Users(ctx web.Context) {
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck

	err := v.orm.Tag("mysql_master").PingContext(ctx.Context())
	if err != nil {
		ctx.ErrorJSON(500, err, web.ErrCtx{"id": id})
		return
	}

	err = v.orm.Tag("sqlite_master").PingContext(ctx.Context())
	if err != nil {
		ctx.ErrorJSON(500, err, web.ErrCtx{"id": id})
		return
	}

	ctx.ErrorJSON(400, fmt.Errorf("user not found"), web.ErrCtx{"id": id})

	logx.Info("user - %d", id)
}
