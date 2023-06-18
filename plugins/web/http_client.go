/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"github.com/osspkg/go-sdk/webutil"
	"github.com/osspkg/goppy/plugins"
)

// WithHTTPClient init pool http clients
func WithHTTPClient() plugins.Plugin {
	return plugins.Plugin{
		Inject: func() ClientHttp {
			return newClientHttp()
		},
	}
}

type (
	ClientHttp interface {
		Create(opts ...webutil.ClientHttpOption) *webutil.ClientHttp
	}

	clientHttp struct {
	}
)

func newClientHttp() ClientHttp {
	return &clientHttp{}
}

func (v *clientHttp) Create(opts ...webutil.ClientHttpOption) *webutil.ClientHttp {
	return webutil.NewClientHttp(opts...)
}
