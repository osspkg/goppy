/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metric

import (
	"reflect"
	"sync/atomic"
	"time"

	"go.osspkg.com/logx"
)

type Writer func(tag, queryName string, execTime time.Duration)

var store atomic.Value

func SetWriter(w Writer) {
	store.Store(w)
}

func ExecTime(tag, name string, call func()) {
	curr := time.Now()
	call()
	since := time.Since(curr)

	if obj := store.Load(); obj != nil {
		if fnw, ok := obj.(Writer); ok {
			fnw(tag, name, since)
		} else {
			logx.Warn("ORM: Invalid metrics writer", "tag", tag, "name", name, "type", reflect.TypeOf(obj).String())
		}
	} else {
		logx.Warn("ORM: Invalid metrics writer", "tag", tag, "name", name, "type", "nil")
	}
}

func init() {
	SetWriter(func(tag, queryName string, execTime time.Duration) {
		logx.Debug("query_time", "tag", tag, "query_name", queryName, "exec_time", execTime.String())
	})
}
