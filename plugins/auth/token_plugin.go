/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"go.osspkg.com/goppy/v3/plugin"
	"go.osspkg.com/goppy/v3/plugins/auth/token"
)

// WithJWT init jwt service
func WithJWT() plugin.Kind {
	return plugin.Kind{
		Config: &token.ConfigGroup{},
		Inject: func(conf *token.ConfigGroup) (token.Token, error) {
			return token.New(conf.JWT)
		},
	}
}
