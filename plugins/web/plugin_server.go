/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import "go.osspkg.com/goppy/v3/plugin"

// WithServer launch of HTTP service with default Router
func WithServer() plugin.Kind {
	return plugin.Kind{
		Config: &ConfigGroup{},
		Inject: func(conf *ConfigGroup) (ServerPool, error) {
			return newServerPool(conf.HTTP)
		},
	}
}
