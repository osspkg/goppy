/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"go.osspkg.com/goppy/v3/plugin"
	"go.osspkg.com/goppy/v3/plugins/auth/oauth"
)

// WithOAuth init oauth service
func WithOAuth(opts ...func(option oauth.Option)) plugin.Kind {
	return plugin.Kind{
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
