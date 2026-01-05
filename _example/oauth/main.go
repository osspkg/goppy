/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v3"
	"go.osspkg.com/goppy/v3/auth"
	"go.osspkg.com/goppy/v3/auth/oauth"
	"go.osspkg.com/goppy/v3/web"
)

func main() {
	app := goppy.New("", "", "")
	app.Plugins(
		web.WithServer(),
		auth.WithOAuth(func(option oauth.Option) {
			option.ApplyProvider(
				&oauth.GoogleProvider{},
				&oauth.YandexProvider{},
			)
		}),
	)
	app.Plugins(
		NewController,
		func(routes web.ServerPool, c *Controller, oa oauth.OAuth) {
			router, ok := routes.Main()
			if !ok {
				return
			}

			router.Use(web.ThrottlingMiddleware(100))

			router.Get("/oauth/r/{code}", oa.Request("code"))
			router.Get("/oauth/c/{code}", oa.Callback("code", c.CallBack))
		},
	)
	app.Run()
}

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) CallBack(ctx web.Ctx, user oauth.User, code oauth.Code) {
	ctx.String(
		200,
		"code: %s, email: %s, name: %s, ico: %s",
		code, user.GetEmail(), user.GetName(), user.GetIcon(),
	)
}
