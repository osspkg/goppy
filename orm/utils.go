/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import "sync/atomic"

type atomicError struct {
	v atomic.Value
}

func (a *atomicError) Set(err error) {
	if err == nil {
		return
	}
	a.v.Store(err)
}

func (a *atomicError) Get() error {
	vv := a.v.Load()
	if vv == nil {
		return nil
	}
	err, ok := vv.(error)
	if !ok {
		return nil
	}
	return err
}
