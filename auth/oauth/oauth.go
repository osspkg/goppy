/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

import (
	"net/http"
	"sync"
)

/**********************************************************************************************************************/

type (
	ConfigItem struct {
		Code         string `yaml:"code"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	}

	Config struct {
		Provider []ConfigItem `yaml:"oauth"`
	}

	configIsp struct {
		State       string
		AuthCodeKey string
		RequestURL  string
	}
)

/**********************************************************************************************************************/

type (
	OAuth struct {
		config *Config
		list   map[string]Provider
		mux    sync.RWMutex
	}

	CallBack func(http.ResponseWriter, *http.Request, User)
)

func New(c *Config) *OAuth {
	return &OAuth{
		config: c,
		list:   make(map[string]Provider),
	}
}

func (v *OAuth) Up() error {
	v.AddProviders(
		&IspYandex{},
	)
	return nil
}

func (v *OAuth) Down() error {
	return nil
}

func (v *OAuth) Request(name string) func(http.ResponseWriter, *http.Request) {
	p, err := v.GetProvider(name)
	if err != nil {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error())) //nolint: errcheck
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, p.AuthCodeURL(), http.StatusMovedPermanently)
	}
}

func (v *OAuth) CallBack(name string, call CallBack) func(w http.ResponseWriter, r *http.Request) {
	p, err := v.GetProvider(name)
	if err != nil {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error())) //nolint: errcheck
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get(p.AuthCodeKey())
		u, err := p.Exchange(r.Context(), code)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error())) //nolint: errcheck
			return
		}
		call(w, r, u)
	}
}
