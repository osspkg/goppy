/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

//go:generate easyjson

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const CodeGoogle = "google"

type (
	//easyjson:json
	modelGoogle struct {
		Name          string `json:"name"`
		Icon          string `json:"picture"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}

	userGoogle struct {
		name  string
		icon  string
		email string
	}
)

func (v *userGoogle) UnmarshalJSON(data []byte) error {
	var tmp modelGoogle
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.EmailVerified {
		v.name = tmp.Name
		v.icon = tmp.Icon
		v.email = tmp.Email
	}

	return nil
}

func (v *userGoogle) GetName() string {
	return v.name
}

func (v *userGoogle) GetIcon() string {
	return v.icon
}

func (v *userGoogle) GetEmail() string {
	return v.email
}

/**********************************************************************************************************************/

type GoogleProvider struct {
	oauth  *oauth2.Config
	config providerConfig
}

func (v *GoogleProvider) Code() string {
	return CodeGoogle
}

func (v *GoogleProvider) Config(c Config) {
	v.oauth = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  c.RedirectURL,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
	v.config = providerConfig{
		State:       "state",
		AuthCodeKey: "code",
		RequestURL:  "https://openidconnect.googleapis.com/v1/userinfo",
	}
}

func (v *GoogleProvider) AuthCodeURL() string {
	return v.oauth.AuthCodeURL(v.config.State)
}

func (v *GoogleProvider) AuthCodeKey() string {
	return v.config.AuthCodeKey
}

func (v *GoogleProvider) Exchange(ctx context.Context, code string) (User, error) {
	m := &userGoogle{}
	if err := oauth2ExchangeContext(ctx, code, v.config.RequestURL, v.oauth, m); err != nil {
		return nil, err
	}
	return m, nil
}
