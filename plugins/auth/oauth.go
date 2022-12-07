package auth

import (
	nethttp "net/http"

	"github.com/dewep-online/goppy/plugins"
	"github.com/dewep-online/goppy/plugins/http"
	"github.com/deweppro/go-auth"
	"github.com/deweppro/go-auth/config"
	"github.com/deweppro/go-auth/providers"
	"github.com/deweppro/go-auth/providers/isp"
	"github.com/deweppro/go-auth/providers/isp/google"
	"github.com/deweppro/go-auth/providers/isp/yandex"
	"github.com/deweppro/go-logger"
)

// ConfigOAuth oauth config model
type ConfigOAuth struct {
	Providers []config.ConfigItem `yaml:"oauth_providers"`
}

func (v *ConfigOAuth) Default() {
	if len(v.Providers) == 0 {
		v.Providers = []config.ConfigItem{
			{
				Code:         google.CODE,
				ClientID:     "****************.apps.googleusercontent.com",
				ClientSecret: "****************",
				RedirectURL:  "https://example.com/oauth/callback/google",
			},
		}
	}
}

// WithOAuth init oauth providers
func WithOAuth() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigOAuth{},
		Inject: func(conf *ConfigOAuth, log logger.Logger) OAuth {
			cc := &config.Config{Provider: make([]config.ConfigItem, 0, len(conf.Providers))}
			cc.Provider = append(cc.Provider, conf.Providers...)
			return &oauthService{
				oa: auth.New(providers.New(cc)),
			}
		},
	}
}

type (
	oauthService struct {
		oa *auth.Auth
	}

	OAuth interface {
		GoogleRequestHandler(ctx http.Ctx)
		GoogleCallbackHandler(handler func(http.Ctx, OAuthUser, ProviderCode)) func(http.Ctx)
		YandexRequestHandler(ctx http.Ctx)
		YandexCallbackHandler(handler func(http.Ctx, OAuthUser, ProviderCode)) func(http.Ctx)
	}

	OAuthUser interface {
		GetName() string
		GetEmail() string
		GetIcon() string
	}

	ProviderCode string
)

func (v *oauthService) requestHandler(code string, ctx http.Ctx) {
	v.oa.Request(code)(ctx.Response(), ctx.Request())
}
func (v *oauthService) callbackHandler(code string, handler func(http.Ctx, OAuthUser, ProviderCode)) func(http.Ctx) {
	return func(ctx http.Ctx) {
		v.oa.CallBack(code, func(_ nethttp.ResponseWriter, _ *nethttp.Request, u isp.IUser) {
			handler(ctx, u, ProviderCode(code))
		})(ctx.Response(), ctx.Request())
	}
}

func (v *oauthService) GoogleRequestHandler(ctx http.Ctx) {
	v.requestHandler(google.CODE, ctx)
}

func (v *oauthService) GoogleCallbackHandler(handler func(http.Ctx, OAuthUser, ProviderCode)) func(http.Ctx) {
	return v.callbackHandler(google.CODE, handler)
}

func (v *oauthService) YandexRequestHandler(ctx http.Ctx) {
	v.oa.Request(yandex.CODE)(ctx.Response(), ctx.Request())
}

func (v *oauthService) YandexCallbackHandler(handler func(http.Ctx, OAuthUser, ProviderCode)) func(http.Ctx) {
	return v.callbackHandler(yandex.CODE, handler)
}
