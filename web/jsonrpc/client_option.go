/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"time"
)

type Opt func(o *cliopts)

type cliopts struct {
	timeout        time.Duration
	keepalive      time.Duration
	genID          func() string
	defaultHeaders map[string]string
	contextHeaders map[string]any
}

func SetGenID(arg func() string) Opt {
	return func(o *cliopts) {
		if arg == nil {
			return
		}
		o.genID = arg
	}
}

func SetTimeout(timeout, keepalive time.Duration) Opt {
	return func(o *cliopts) {
		o.timeout = max(timeout, time.Second)
		o.keepalive = max(keepalive, time.Second)
	}
}

func SetHeader(key, value string) Opt {
	return func(o *cliopts) {
		o.defaultHeaders[key] = value
	}
}

func SetContextHeader(header string, key any) Opt {
	return func(o *cliopts) {
		o.contextHeaders[header] = key
	}
}
