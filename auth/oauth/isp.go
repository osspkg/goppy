/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/ioutil"
	"golang.org/x/oauth2"
)

var (
	errProviderFail = errors.New("provider not found")
)

type (
	User interface {
		GetName() string
		GetEmail() string
		GetIcon() string
	}

	Provider interface {
		Code() string
		Config(conf ConfigItem)
		AuthCodeURL() string
		AuthCodeKey() string
		Exchange(ctx context.Context, code string) (User, error)
	}
)

func (v *OAuth) AddProviders(p ...Provider) {
	v.mux.Lock()
	defer v.mux.Unlock()

	for _, item := range p {
		for _, cp := range v.config.Provider {
			if cp.Code == item.Code() {
				item.Config(cp)
				v.list[item.Code()] = item
			}
		}
	}
}

func (v *OAuth) GetProvider(name string) (Provider, error) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	p, ok := v.list[name]
	if !ok {
		return nil, errProviderFail
	}
	return p, nil
}

/**********************************************************************************************************************/

type oauth2Config interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

func oauth2ExchangeContext(
	ctx context.Context, code string, uri string, srv oauth2Config, model json.Unmarshaler,
) error {
	tok, err := srv.Exchange(ctx, code)
	if err != nil {
		return errors.Wrapf(err, "exchange to oauth service")
	}
	client := srv.Client(ctx, tok)
	resp, err := client.Get(uri) //nolint: bodyclose
	if err != nil {
		return errors.Wrapf(err, "client request to oauth service")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "read response from oauth service")
	}
	if err = json.Unmarshal(b, model); err != nil {
		return errors.Wrapf(err, "decode oauth model")
	}
	return nil
}
