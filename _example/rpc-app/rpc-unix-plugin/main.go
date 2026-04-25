/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.osspkg.com/logx"

	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/_example/rpc-app/rpc-unix-plugin/transport"
	"go.osspkg.com/goppy/v3/web"
	"go.osspkg.com/goppy/v3/web/jsonrpc"
)

func main() {
	app := goppy.New("app_name", "v1.0.0", "app description")
	app.Plugins(
		web.WithServer(),
		jsonrpc.WithTransport(
			jsonrpc.Path("/"),
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
			t.Add(transport.NewJSONRPCApiTransport(c, []string{"main"}))
			return nil
		},
	)
	app.Run()
}

type Controller struct{}

func (c Controller) Ping(ctx context.Context) (pong bool, err error) {
	pong = true
	go func() {
		time.Sleep(time.Second)
		os.Exit(0)
	}()
	return
}

func NewController() *Controller { return new(Controller) }
