/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"go.osspkg.com/goppy/v2/auth/oauth"
	"go.osspkg.com/goppy/v2/plugins"
)

// WithOAuth init oauth service
func WithOAuth(opts ...func(option oauth.Option)) plugins.Plugin {
	return plugins.Plugin{
		Config: &oauth.ConfigGroup{},
		Inject: func(conf *oauth.ConfigGroup) oauth.OAuth {
			obj := oauth.New(conf.Providers)
			for _, opt := range opts {
				opt(obj)
			}
			return obj
		},
	}
}
