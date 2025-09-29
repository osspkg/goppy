/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"net/http"
	"net/url"
)

func proxyUrl(proxy string) func(r *http.Request) (*url.URL, error) {
	if len(proxy) == 0 || proxy == "env" {
		return http.ProxyFromEnvironment
	}

	u, err := url.Parse(proxy)
	if err != nil {
		return func(r *http.Request) (*url.URL, error) {
			return nil, err
		}
	}

	return http.ProxyURL(u)
}
