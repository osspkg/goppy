package main

import (
	"github.com/osspkg/goppy"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/auth"
	"github.com/osspkg/goppy/plugins/web"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
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
