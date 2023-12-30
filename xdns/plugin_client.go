/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xdns

import "go.osspkg.com/goppy/plugins"

func WithDNSClient(opts ...ClientOption) plugins.Plugin {
	return plugins.Plugin{
		Inject: func() *Client {
			return NewClient(opts...)
		},
	}
}
