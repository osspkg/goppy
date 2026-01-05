/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

type (
	ConfigGroup struct {
		Providers []Config `yaml:"oauth"`
	}

	Config struct {
		Code         string `yaml:"code"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURL  string `yaml:"redirect_url"`
	}
)

func (v *ConfigGroup) Default() {
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
