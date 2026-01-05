/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"net/http"
	"time"

	"go.osspkg.com/goppy/v3/auth/signature"
	"go.osspkg.com/goppy/v3/web/client/comparison"
)

type HTTPOption func(c *httpCli)

func WithProxy(proxy string) HTTPOption {
	return func(c *httpCli) {
		c.nativeClient.Transport.(*http.Transport).Proxy = proxyUrl(proxy)
	}
}

func WithTimeouts(timeout, keepAlive time.Duration) HTTPOption {
	return func(c *httpCli) {
		c.netDialer.Timeout = timeout
		c.netDialer.KeepAlive = keepAlive
	}
}

func WithMaxIdleConns(count int) HTTPOption {
	return func(c *httpCli) {
		c.nativeClient.Transport.(*http.Transport).MaxIdleConns = count
	}
}

func WithDefaultHeaders(h map[string]string) HTTPOption {
	return func(c *httpCli) {
		c.defaultHeaders = make(http.Header, len(h))
		for k, v := range h {
			c.defaultHeaders.Set(k, v)
		}
	}
}

func WithComparisonType(types ...comparison.Type) HTTPOption {
	return func(c *httpCli) {
		c.types = types
	}
}

func WithSignatures(sigs map[string]signature.Signature) HTTPOption {
	return func(c *httpCli) {
		c.signStore.Replace(sigs)
	}
}
