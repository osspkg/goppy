package main

import (
	"github.com/dewep-online/goppy"
	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/auth"
	"github.com/dewep-online/goppy/plugins/http"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml")
	app.Plugins(
		http.WithHTTP(),
		auth.WithOAuth(),
	)
	app.Plugins(
		plugins.Plugin{
			Inject: NewController,
			Resolve: func(routes http.RouterPool, c *Controller, oa auth.OAuth) {
				router := routes.Main()
				router.Use(http.ThrottlingMiddleware(100))

				router.Get("/oauth/r/google", oa.GoogleRequestHandler)
				router.Get("/oauth/c/google", oa.GoogleCallbackHandler(c.CallBack))

				router.Get("/oauth/r/yandex", oa.YandexRequestHandler)
				router.Get("/oauth/c/yandex", oa.YandexCallbackHandler(c.CallBack))
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

func (v *Controller) CallBack(ctx http.Ctx, user auth.OAuthUser, code auth.ProviderCode) {
	ctx.SetBody(200).
		String("code: %s, email: %s, name: %s, ico: %s", code, user.GetEmail(), user.GetName(), user.GetIcon())
}
