/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xc

import cc "context"

type (
	_ctx struct {
		ctx    cc.Context
		cancel cc.CancelFunc
	}

	Context interface {
		Close()
		Context() cc.Context
		Done() <-chan struct{}
	}
)

func New() Context {
	ctx, cancel := cc.WithCancel(cc.Background())
	return &_ctx{
		ctx:    ctx,
		cancel: cancel,
	}
}

func NewContext(c cc.Context) Context {
	ctx, cancel := cc.WithCancel(c)
	return &_ctx{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Close context close method
func (v *_ctx) Close() {
	v.cancel()
}

// Context general context
func (v *_ctx) Context() cc.Context {
	return v.ctx
}

// Done context close wait channel
func (v *_ctx) Done() <-chan struct{} {
	return v.ctx.Done()
}
