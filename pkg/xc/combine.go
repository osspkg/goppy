/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xc

import (
	"context"
	"errors"
	"reflect"
	"time"
)

type joinedCtx struct {
	main  context.Context
	multi []context.Context
}

func (j joinedCtx) Deadline() (deadline time.Time, has bool) {
	for _, ctx := range j.multi {
		dl, ok := ctx.Deadline()
		if !ok {
			continue
		}

		if !has {
			deadline, has = dl, true
			continue
		}

		if dl.Before(deadline) {
			deadline = dl
		}
	}

	return
}

func (j joinedCtx) Done() <-chan struct{} {
	return j.main.Done()
}

func (j joinedCtx) Err() (err error) {
	for _, ctx := range j.multi {
		err = errors.Join(err, ctx.Err())
	}
	return
}

func (j joinedCtx) Value(key any) any {
	for _, ctx := range j.multi {
		if value := ctx.Value(key); value != nil {
			return value
		}
	}
	return nil
}

func (joinedCtx) String() string {
	return "xc.Join"
}

func Join(multi ...context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	if len(multi) == 0 {
		return ctx, cancel
	}

	multi = append(multi, ctx)

	jCtx := joinedCtx{
		main:  ctx,
		multi: multi,
	}

	startC := make(chan struct{}, 1)

	go func() {
		cases := make([]reflect.SelectCase, 0, len(multi))
		for _, vv := range multi {
			cases = append(cases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(vv.Done()),
			})
		}

		close(startC)

		chosen, _, _ := reflect.Select(cases)
		switch chosen {
		default:
			cancel()
		}
	}()

	<-startC

	return jCtx, cancel
}
