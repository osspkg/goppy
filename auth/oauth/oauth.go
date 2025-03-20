/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils"
	"go.osspkg.com/logx"
	"golang.org/x/oauth2"

	"go.osspkg.com/goppy/v2/web"
)

type (
	OAuth interface {
		ApplyProvider(p ...Provider)
		Request(code string) func(web.Context)
		Callback(code string, handler func(web.Context, User, Code)) func(web.Context)
	}

	Option interface {
		ApplyProvider(p ...Provider)
	}

	User interface {
		GetName() string
		GetEmail() string
		GetIcon() string
	}

	Provider interface {
		Code() string
		Config(conf Config)
		AuthCodeURL() string
		AuthCodeKey() string
		Exchange(ctx context.Context, code string) (User, error)
	}

	Code string
)

type (
	service struct {
		config []Config
		list   map[string]Provider
	}

	providerConfig struct {
		State       string
		AuthCodeKey string
		RequestURL  string
	}
)

func New(c []Config) OAuth {
	return &service{
		config: c,
		list:   make(map[string]Provider),
	}
}

func (v *service) ApplyProvider(p ...Provider) {
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

func (v *service) getProvider(name string) (Provider, error) {
	if p, ok := v.list[name]; ok {
		return p, nil
	}

	return nil, ErrProviderNotFound
}

func (v *service) Request(code string) func(web.Context) {
	return func(ctx web.Context) {
		name, err := ctx.Param(code).String()
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]any{
				code: name,
			})
			return
		}

		p, err := v.getProvider(name)
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]any{})
			return
		}

		ctx.Redirect(p.AuthCodeURL())
	}
}

func (v *service) Callback(code string, handler func(web.Context, User, Code)) func(web.Context) {
	return func(ctx web.Context) {
		name, err := ctx.Param(code).String()
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]any{
				code: name,
			})
			return
		}

		p, err := v.getProvider(name)
		if err != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err, map[string]any{})
			return
		}

		u, err0 := p.Exchange(ctx.Context(), ctx.Query(p.AuthCodeKey()))
		if err0 != nil {
			ctx.ErrorJSON(http.StatusBadRequest, err0, map[string]any{})
			return
		}

		handler(ctx, u, Code(name))
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
