/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v2/orm/clients/sqlite"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/orm"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
)

func main() {

	app := goppy.New("goppy_database", "v1.0.0", "")
	app.Plugins(
		web.WithServer(),
		orm.WithORM(sqlite.Name),
		orm.WithMigration(orm.Migration{
			Tags:    []string{"sqlite_master"},
			Dialect: "mysql",
			Data: map[string]string{
				"0002_data.sql": `
					CREATE TABLE IF NOT EXISTS "demo2"
					(
						"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT
					);
`,
			},
		}),
		orm.WithMigration(),
	)
	app.Plugins(
		plugins.Kind{
			Inject: NewController,
			Resolve: func(routes web.ServerPool, c *Controller) {
				router, ok := routes.Main()
				if !ok {
					return
				}

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

func (v *Controller) Users(ctx web.Ctx) {
	data := Model{data: []int64{1, 2, 3, 4}}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Ctx) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck

	err := v.orm.Tag("sqlite_master").PingContext(ctx.Context())
	if err != nil {
		ctx.ErrorJSON(500, err, "id", id)
		return
	}

	ctx.ErrorJSON(200, fmt.Errorf("user not found"), "id", id)

	logx.Info("user", "id", id)
}

type Model struct {
	data []int64
}

func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}
