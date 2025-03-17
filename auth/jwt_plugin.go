/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package auth

import (
	"go.osspkg.com/goppy/v2/auth/jwt"
	"go.osspkg.com/goppy/v2/plugins"
)

// WithJWT init jwt service
func WithJWT() plugins.Plugin {
	return plugins.Plugin{
		Config: &jwt.ConfigGroup{},
		Inject: func(conf *jwt.ConfigGroup) jwt.JWT {
			return jwt.New(conf.JWT)
		},
	}
}
