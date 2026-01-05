/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v3/console"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/metrics"
	"go.osspkg.com/goppy/v3/web"
)

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		metrics.WithServer(),
		web.WithServer(),
	)
	app.Plugins(
		NewController,
		func(routes web.ServerPool, c *Controller) {
			router, ok := routes.Main()
			if !ok {
				return
			}

			router.Use(web.ThrottlingMiddleware(100))
			router.Get("/users", c.Users)

			api := router.Collection("/api/v1", web.ThrottlingMiddleware(100))
			api.Get("/user/{id}", c.User)
		},
	)
	app.Command(func(setter console.CommandSetter) {
		setter.Setup("env", "show all envs")
		setter.ExecFunc(func() {
			fmt.Println(os.Environ())
		})
	})
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) Users(ctx web.Ctx) {
	metrics.Gauge("users_request").Inc()
	data := Model{
		data: []int64{1, 2, 3, 4},
	}
	ctx.JSON(200, data)
}

func (v *Controller) User(ctx web.Ctx) {
	id, _ := ctx.Param("id").Int() // nolint: errcheck
	ctx.String(200, "user id: %d", id)
	logx.Info("user - %d", id)
}

type Model struct {
	data []int64
}

func (m Model) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.data)
}
