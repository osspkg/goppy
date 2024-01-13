/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import "go.osspkg.com/goppy/plugins"

// WithHTTPClient init pool http clients
func WithHTTPClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func() ClientHttpPool {
			return newClientHttp()
		},
	}
}

type (
	ClientHttpPool interface {
		Create(opts ...ClientHttpOption) *ClientHttp
	}

	clientHttp struct {
	}
)

func newClientHttp() ClientHttpPool {
	return &clientHttp{}
}

func (v *clientHttp) Create(opts ...ClientHttpOption) *ClientHttp {
	return NewClientHttp(opts...)
}
