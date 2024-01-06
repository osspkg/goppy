/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/auth"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/web"
)

func main() {
	app := goppy.New()
	app.Plugins(
		web.WithHTTP(),
		auth.WithOAuth(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes web.RouterPool, c *Controller, oa auth.OAuth) {
				router := routes.Main()
				router.Use(web.ThrottlingMiddleware(100))

				router.Get("/oauth/r/{code}", oa.RequestHandler("code"))
				router.Get("/oauth/c/{code}", oa.CallbackHandler("code", c.CallBack))
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

func (v *Controller) CallBack(ctx web.Context, user auth.OAuthUser, code auth.Code) {
	ctx.String(200, "code: %s, email: %s, name: %s, ico: %s", code, user.GetEmail(), user.GetName(), user.GetIcon())
}
