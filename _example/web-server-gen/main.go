/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"
	"time"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/_example/web-server-gen/transport"
	"go.osspkg.com/goppy/v3/_example/web-server-gen/types"
	"go.osspkg.com/goppy/v3/web"
	"go.osspkg.com/goppy/v3/web/jsonrpc"
)

func main() {
	// Specify the path to the config via the argument: `--config`.
	// Specify the path to the pidfile via the argument: `--pid`.
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		web.WithServer(),
		jsonrpc.WithTransport(
			jsonrpc.Path("/rpc"),
			jsonrpc.Timeout(5*time.Second),
			jsonrpc.ErrHandler(func(method string, err error) error {
				logx.Error("json-rpc call failed", "method", method, "err", err)
				return fmt.Errorf("json-rpc call failed: %w", err)
			}),
		),
	)
	app.Plugins(
		NewController,
		func(t jsonrpc.Transport, c *Controller) error {
			t.Inject(transport.NewJSONRPCApiTransport(c, []string{"main"}))
			t.Inject(transport.NewJSONRPCUserTransport(c, []string{"main"}))
			t.Inject(transport.NewJSONRPCPostTransport(c, []string{"main"}))
			return nil
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
