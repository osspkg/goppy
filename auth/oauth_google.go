/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

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

	oauthUserGoogle struct {
		name  string
		icon  string
		email string
	}
)

func (v *oauthUserGoogle) UnmarshalJSON(data []byte) error {
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

func (v *oauthUserGoogle) GetName() string {
	return v.name
}

func (v *oauthUserGoogle) GetIcon() string {
	return v.icon
}

func (v *oauthUserGoogle) GetEmail() string {
	return v.email
}

/**********************************************************************************************************************/

type OAuthGoogleProvider struct {
	oauth  *oauth2.Config
	config oauthProviderConfig
}

func (v *OAuthGoogleProvider) Code() string {
	return CodeGoogle
}

func (v *OAuthGoogleProvider) Config(c Config) {
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
	v.config = oauthProviderConfig{
		State:       "state",
		AuthCodeKey: "code",
		RequestURL:  "https://openidconnect.googleapis.com/v1/userinfo",
	}
}

func (v *OAuthGoogleProvider) AuthCodeURL() string {
	return v.oauth.AuthCodeURL(v.config.State)
}

func (v *OAuthGoogleProvider) AuthCodeKey() string {
	return v.config.AuthCodeKey
}

func (v *OAuthGoogleProvider) Exchange(ctx context.Context, code string) (OAuthUser, error) {
	m := &oauthUserGoogle{}
	if err := oauth2ExchangeContext(ctx, code, v.config.RequestURL, v.oauth, m); err != nil {
		return nil, err
	}
	return m, nil
}
