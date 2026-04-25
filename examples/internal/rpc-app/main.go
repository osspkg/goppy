/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/rpc"
	"go.osspkg.com/goppy/v3/web"
)

func main() {
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		web.WithServer(),
		rpc.WithRPC(),
	)
	app.Plugins(
		NewController,
		func(routes web.ServerPool, c *Controller) {
			router, ok := routes.Main()
			if !ok {
				return
			}

			router.Use(web.ThrottlingMiddleware(100))
			router.Get("/call/{app}", c.Call)
		},
	)
	app.Run()
}

type Controller struct {
	rpc *rpc.RPC
}

func NewController(rpc *rpc.RPC) *Controller {
	return &Controller{rpc: rpc}
}

func (c *Controller) Call(ctx web.Ctx) {
	appName, _ := ctx.Param("app").String()

	var result any

	err := c.rpc.Call(ctx.Context(), appName, "api.ping", struct{}{}, &result)
	if err != nil {
		ctx.ErrorJSON(400, err)
	} else {
		ctx.JSON(200, result)
	}
}
