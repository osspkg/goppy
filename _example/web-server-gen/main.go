/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/_example/web-server-gen/transport"
	"go.osspkg.com/goppy/v3/_example/web-server-gen/types"
	"go.osspkg.com/goppy/v3/web"
)

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		web.WithServer(),
	)
	app.Plugins(
		NewController,
		func(routes web.ServerPool, c *Controller) *transport.JSONRPCHandler {
			return transport.NewJSONRPCHandler(routes, c, c, c)
		},
	)
	app.Run()
}

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c Controller) Name(ctx context.Context, userID int64) (name string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) ByID(ctx context.Context, ID int64) (text bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) List(ctx context.Context, userID int64) (text []types.Text, err error) {
	//TODO implement me
	panic("implement me")
}

func (c Controller) Root(ctx context.Context, userID int64, userName string) (status bool, err error) {
	switch userID {
	case 0:
		return false, fmt.Errorf("userID 0")
	default:
		return true, nil
	}
}

func (c Controller) Auth(ctx context.Context, userID int64, userName string) (status bool, err error) {
	switch userID {
	case 0:
		return false, fmt.Errorf("userID 0")
	default:
		return true, nil
	}
}
