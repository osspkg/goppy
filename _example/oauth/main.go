/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/auth"
	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
)

func main() {
	app := goppy.New("", "", "")
	app.Plugins(
		web.WithServer(),
		auth.WithOAuth(func(option auth.OAuthOption) {
			option.ApplyProvider(
				&auth.OAuthGoogleProvider{},
				&auth.OAuthYandexProvider{},
			)
		}),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes web.RouterPool, c *Controller, oa auth.OAuth) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))

				router.Get("/oauth/r/{code}", oa.Request("code"))
				router.Get("/oauth/c/{code}", oa.Callback("code", c.CallBack))
			},
		},
	)
	app.Run()
}

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (v *Controller) CallBack(ctx web.Context, user auth.OAuthUser, code auth.OAuthCode) {
	ctx.String(
		200,
		"code: %s, email: %s, name: %s, ico: %s",
		code, user.GetEmail(), user.GetName(), user.GetIcon(),
	)
}
