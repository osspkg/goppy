/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.osspkg.com/console"

	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/_example/json-rpc-server/models"
	"go.osspkg.com/goppy/v2/metrics"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
	"go.osspkg.com/goppy/v2/web/jsonrpc"
)

func main() {
	app := goppy.New("app_name", "v1.0.0", "app description")
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

				rpc := jsonrpc.NewCaller()

				fmt.Println(rpc.Add("base.user", c.User))
				fmt.Println(rpc.Add("base.users", c.Users))
				fmt.Println(rpc.Add("base.proxy", c.Proxy))

				router.Post("/rpc", rpc.Handler)
			},
		},
	)
	app.Command("env", func(s console.CommandSetter) {
		fmt.Println(os.Environ())
	})
	app.Run()
}

type Controller struct {
	cli *jsonrpc.Client
}

func NewController() *Controller {
	return &Controller{
		cli: jsonrpc.NewClient("http://localhost:10000/rpc"),
	}
}

func (v *Controller) Users(ctx context.Context, ids *models.IntArray) (*models.Users, error) {
	result := make(models.Users, 0, len(*ids))
	for _, id := range *ids {
		result = append(result, models.User{Id: id, Name: fmt.Sprintf("user #%d", id)})
	}
	return &result, nil
}

func (v *Controller) User(ctx context.Context, ids *models.IntArray) (*models.User, error) {
	if len(*ids) != 1 {
		return nil, fmt.Errorf("must has one element")
	}

	return &models.User{Id: (*ids)[0], Name: fmt.Sprintf("user #%d", (*ids)[0])}, nil
}

func (v *Controller) Proxy(ctx context.Context, ids *models.IntArray) (*models.Users, error) {
	result := make(models.Users, 0, len(*ids))
	batchs := make([]jsonrpc.Batch, 0, len(*ids))
	for _, id := range *ids {
		params := models.IntArray([]int{id})
		batchs = append(batchs, jsonrpc.Batch{
			Method: "base.users",
			Params: &params,
			Fallback: func(d json.RawMessage, _ *jsonrpc.Error) {
				var us models.Users
				json.Unmarshal(d, &us)
				for _, u := range us {
					u.Name = "PROXY: " + u.Name
					result = append(result, u)
				}
			},
		})
	}
	if err := v.cli.Call(ctx, batchs...); err != nil {
		return nil, err
	}
	return &result, nil
}
