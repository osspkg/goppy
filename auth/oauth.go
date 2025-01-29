/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils"
	"go.osspkg.com/logx"
	"golang.org/x/oauth2"

	"go.osspkg.com/goppy/v2/plugins"
	"go.osspkg.com/goppy/v2/web"
)

var (
	errProviderFail = errors.New("provider not found")
)

type (
	// ConfigOAuth oauth config model
	ConfigOAuth struct {
		Providers []Config `yaml:"oauth"`
	}

	Config struct {
		Code         string `yaml:"code"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	}
)

func (v *ConfigOAuth) Default() {
	if len(v.Providers) == 0 {
		v.Providers = []Config{
			{
				Code:         CodeGoogle,
				ClientID:     "****************.apps.googleusercontent.com",
				ClientSecret: "****************",
				RedirectURL:  "https://example.com/oauth/callback/google",
			},
		}
	}
}

// WithOAuth init oauth providers
func WithOAuth(opts ...func(OAuthOption)) plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigOAuth{},
		Inject: func(conf *ConfigOAuth) OAuth {
			obj := &oauthService{
				config: conf.Providers,
				list:   make(map[string]OAuthProvider),
			}

			for _, opt := range opts {
				opt(obj)
			}

			return obj
		},
	}
}

type (
	oauthService struct {
		config []Config
		list   map[string]OAuthProvider
	}

	oauthProviderConfig struct {
		State       string
		AuthCodeKey string
		RequestURL  string
	}

	OAuth interface {
		Request(code string) func(web.Context)
		Callback(code string, handler func(web.Context, OAuthUser, OAuthCode)) func(web.Context)
	}

	OAuthOption interface {
		ApplyProvider(p ...OAuthProvider)
	}

	OAuthUser interface {
		GetName() string
		GetEmail() string
		GetIcon() string
	}

	OAuthProvider interface {
		Code() string
		Config(conf Config)
		AuthCodeURL() string
		AuthCodeKey() string
		Exchange(ctx context.Context, code string) (OAuthUser, error)
	}

	OAuthCode string
)

func (v *oauthService) ApplyProvider(p ...OAuthProvider) {
	for _, item := range p {
		for _, cp := range v.config {
			if cp.Code == item.Code() {
				logx.Info("OAuth add provider", "name", cp.Code)
				item.Config(cp)
				v.list[item.Code()] = item
			}
		}
	}
}

func (v *oauthService) getProvider(name string) (OAuthProvider, error) {
	p, ok := v.list[name]
	if !ok {
		return nil, errProviderFail
	}
	return p, nil
}

func (v *oauthService) Request(code string) func(web.Context) {
	return func(ctx web.Context) {
		name, err := ctx.Param(code).String()
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]interface{}{
				code: name,
			})
			return
		}
		p, err := v.getProvider(name)
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]interface{}{})
			return
		}
		ctx.Redirect(p.AuthCodeURL())
	}
}

func (v *oauthService) Callback(code string, handler func(web.Context, OAuthUser, OAuthCode)) func(web.Context) {
	return func(ctx web.Context) {
		name, err := ctx.Param(code).String()
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]interface{}{
				code: name,
			})
			return
		}
		p, err := v.getProvider(name)
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]interface{}{})
			return
		}
		u, err0 := p.Exchange(ctx.Context(), ctx.Query(p.AuthCodeKey()))
		if err0 != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err0, map[string]interface{}{})
			return
		}
		handler(ctx, u, OAuthCode(name))
	}
}

type oauth2Config interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

func oauth2ExchangeContext(ctx context.Context, code, uri string, srv oauth2Config, model json.Unmarshaler) error {
	tok, err := srv.Exchange(ctx, code)
	if err != nil {
		return errors.Wrapf(err, "exchange to oauth service")
	}
	client := srv.Client(ctx, tok)
	resp, err := client.Get(uri) // nolint: bodyclose
	if err != nil {
		return errors.Wrapf(err, "client request to oauth service")
	}
	b, err := ioutils.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "read response from oauth service")
	}
	if err = json.Unmarshal(b, model); err != nil {
		return errors.Wrapf(err, "decode oauth model")
	}
	return nil
}
