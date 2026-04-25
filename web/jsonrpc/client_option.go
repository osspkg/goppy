/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import (
	"time"

	"go.osspkg.com/goppy/v3/web/client"
)

type Opt func(o *cliopts)

type cliopts struct {
	genID    func() string
	httpopts []client.HTTPOption
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
		o.httpopts = append(o.httpopts, client.WithTimeouts(timeout, keepalive))
	}
}

func SetHeader(key, value string) Opt {
	return func(o *cliopts) {
		o.httpopts = append(o.httpopts, client.WithStaticHeader(key, value))
	}
}

func SetContextHeader(header string, key any) Opt {
	return func(o *cliopts) {
		o.httpopts = append(o.httpopts, client.WithContextHeaderValue(header, key))
	}
}

func SetUnixSocket(path string) Opt {
	return func(o *cliopts) {
		o.httpopts = append(o.httpopts, client.WithUnixSocket(path))
	}
}
