/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package appsteps

import (
	"context"
	"strings"
)

type item struct {
	cancel context.CancelFunc
	ctx    context.Context
}

type Step struct {
	data map[string]item
}

func New(global context.Context, names ...string) *Step {
	obj := &Step{
		data: make(map[string]item),
	}

	for _, name := range names {
		if _, ok := obj.data[name]; ok {
			panic("duplicate step name: " + name)
		}

		var (
			ctx    context.Context
			cancel context.CancelFunc
		)

		if strings.HasPrefix(name, "*") {
			ctx, cancel = context.WithCancel(context.Background())
		} else {
			ctx, cancel = context.WithCancel(global)
		}

		obj.data[name] = item{
			cancel: cancel,
			ctx:    ctx,
		}
	}

	return obj
}

func (s *Step) Done(name string) *Step {
	if c, ok := s.data[name]; ok {
		c.cancel()
	} else {
		panic("no such step: " + name)
	}
	return s
}

func (s *Step) Wait(name string) *Step {
	if c, ok := s.data[name]; ok {
		<-c.ctx.Done()
	} else {
		panic("no such step: " + name)
	}
	return s
}
