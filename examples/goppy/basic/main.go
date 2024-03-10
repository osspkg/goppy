/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"fmt"
	"os"

	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/console"
	"go.osspkg.com/goppy/metrics"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/web"
)

func main() {
	app := goppy.New()
	app.AppName("goppy_base_app")
	app.AppVersion("v1.0.0")
	app.Plugins(
		metrics.WithServer(),
		web.WithServer(),
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
	app.Command("env", func(s console.CommandSetter) {
		fmt.Println(os.Environ())
	})
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx web.Context) {
	metrics.Gauge("users_request").Inc()
	data := []int64{1, 2, 3, 4}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Context) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck
	ctx.String(200, "user id: %d", id)
	ctx.Log().Infof("user - %d", id)
}
