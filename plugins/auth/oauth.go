/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"net/http"

	"github.com/osspkg/go-sdk/auth/oauth"
	"github.com/osspkg/goppy/plugins"
	"github.com/osspkg/goppy/plugins/web"
)

// ConfigOAuth oauth config model
type ConfigOAuth struct {
	Providers []oauth.ConfigItem `yaml:"oauth"`
}

func (v *ConfigOAuth) Default() {
	if len(v.Providers) == 0 {
		v.Providers = []oauth.ConfigItem{
			{
				Code:         oauth.CodeGoogle,
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
		Inject: func(conf *ConfigOAuth) OAuth {
			cc := &oauth.Config{Provider: make([]oauth.ConfigItem, 0, len(conf.Providers))}
			cc.Provider = append(cc.Provider, conf.Providers...)
			return &oauthService{
				oa: oauth.New(cc),
			}
		},
	}
}

type (
	oauthService struct {
		oa *oauth.OAuth
	}

	OAuth interface {
		RequestHandler(code string) func(web.Context)
		CallbackHandler(code string, handler func(web.Context, OAuthUser, Code)) func(web.Context)
	}

	OAuthUser interface {
		GetName() string
		GetEmail() string
		GetIcon() string
	}

	Code string
)

func (v *oauthService) RequestHandler(code string) func(web.Context) {
	return func(ctx web.Context) {
		val, err := ctx.Param(code).String()
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]interface{}{
				code: val,
			})
			return
		}
		v.oa.Request(val)(ctx.Response(), ctx.Request())
	}
}

func (v *oauthService) CallbackHandler(code string, handler func(web.Context, OAuthUser, Code)) func(web.Context) {
	return func(ctx web.Context) {
		v.oa.CallBack(code, func(_ http.ResponseWriter, _ *http.Request, u oauth.User) {
			handler(ctx, u, Code(code))
		})(ctx.Response(), ctx.Request())
	}
}
