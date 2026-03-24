/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package jsonrpc

import "time"

type Option func(o *options)

type options struct {
	timeout    time.Duration
	path       string
	errHandler func(method string, err error) error
}

func Timeout(arg time.Duration) Option {
	return func(o *options) {
		if arg <= time.Second {
			arg = time.Second
		}
		o.timeout = arg
	}
}

func Path(arg string) Option {
	return func(o *options) {
		if len(arg) == 0 {
			arg = "/"
		}
		o.path = arg
	}
}

func ErrHandler(arg func(method string, err error) error) Option {
	return func(o *options) {
		if arg == nil {
			arg = func(_ string, err error) error {
				return err
			}
		}
		o.errHandler = arg
	}
}
