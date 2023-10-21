/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package routine

import (
	"context"
	"time"

	"github.com/osspkg/goppy/sdk/errors"
	"github.com/osspkg/goppy/sdk/iosync"
)

func Interval(ctx context.Context, interval time.Duration, call func(context.Context)) {
	call(ctx)

	go func() {
		tick := time.NewTicker(interval)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				call(ctx)
			}
		}
	}()
}

func Retry(count int, ttl time.Duration, call func() error) error {
	var err error
	for i := 0; i < count; i++ {
		if e := call(); e != nil {
			err = errors.Wrap(err, errors.Wrapf(e, "[#%d]", i))
			time.Sleep(ttl)
			continue
		}
		return nil
	}
	return errors.Wrapf(err, "retry error")
}

func Parallel(calls ...func()) {
	wg := iosync.NewGroup()
	for _, call := range calls {
		call := call
		wg.Background(func() {
			call()
		})
	}
	wg.Wait()
}
