/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

//go:generate easyjson

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

const CodeYandex = "yandex"

type (
	//easyjson:json
	modelYandex struct {
		Name  string `json:"display_name"`
		Icon  string `json:"default_avatar_id"`
		Email string `json:"default_email"`
	}

	userYandex struct {
		name  string
		icon  string
		email string
	}
)

func (v *userYandex) UnmarshalJSON(data []byte) error {
	var tmp modelYandex
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if len(tmp.Icon) > 0 {
		v.icon = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-retina-50", tmp.Icon)
	}
	v.name = tmp.Name
	v.email = tmp.Email

	return nil
}

func (v *userYandex) GetName() string {
	return v.name
}

func (v *userYandex) GetIcon() string {
	return v.icon
}

func (v *userYandex) GetEmail() string {
	return v.email
}

/**********************************************************************************************************************/

type YandexProvider struct {
	oauth  *oauth2.Config
	config providerConfig
}

func (v *YandexProvider) Code() string {
	return CodeYandex
}

func (v *YandexProvider) Config(c Config) {
	v.oauth = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Endpoint:     yandex.Endpoint,
		Scopes: []string{
			"login:email",
			"login:info",
			"login:avatar",
		},
	}
	v.config = providerConfig{
		State:       "state",
		AuthCodeKey: "code",
		RequestURL:  "https://login.yandex.ru/info",
	}
}

func (v *YandexProvider) AuthCodeURL() string {
	return v.oauth.AuthCodeURL(v.config.State)
}

func (v *YandexProvider) AuthCodeKey() string {
	return v.config.AuthCodeKey
}

func (v *YandexProvider) Exchange(ctx context.Context, code string) (User, error) {
	m := &userYandex{}
	if err := oauth2ExchangeContext(ctx, code, v.config.RequestURL, v.oauth, m); err != nil {
		return nil, err
	}
	return m, nil
}
