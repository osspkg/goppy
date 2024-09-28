/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

//go:generate easyjson

const CodeYandex = "yandex"

type (
	//easyjson:json
	modelYandex struct {
		Name  string `json:"display_name"`
		Icon  string `json:"default_avatar_id"`
		Email string `json:"default_email"`
	}

	oauthUserYandex struct {
		name  string
		icon  string
		email string
	}
)

func (v *oauthUserYandex) UnmarshalJSON(data []byte) error {
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

func (v *oauthUserYandex) GetName() string {
	return v.name
}

func (v *oauthUserYandex) GetIcon() string {
	return v.icon
}

func (v *oauthUserYandex) GetEmail() string {
	return v.email
}

/**********************************************************************************************************************/

type OAuthYandexProvider struct {
	oauth  *oauth2.Config
	config oauthProviderConfig
}

func (v *OAuthYandexProvider) Code() string {
	return CodeYandex
}

func (v *OAuthYandexProvider) Config(c Config) {
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
	v.config = oauthProviderConfig{
		State:       "state",
		AuthCodeKey: "code",
		RequestURL:  "https://login.yandex.ru/info",
	}
}

func (v *OAuthYandexProvider) AuthCodeURL() string {
	return v.oauth.AuthCodeURL(v.config.State)
}

func (v *OAuthYandexProvider) AuthCodeKey() string {
	return v.config.AuthCodeKey
}

func (v *OAuthYandexProvider) Exchange(ctx context.Context, code string) (OAuthUser, error) {
	m := &oauthUserYandex{}
	if err := oauth2ExchangeContext(ctx, code, v.config.RequestURL, v.oauth, m); err != nil {
		return nil, err
	}
	return m, nil
}
