/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

var debugStatus atomic.Bool

func ShowDebug(v bool) {
	debugStatus.Store(v)
}

func dbg(level int, message string, args ...any) {
	if !debugStatus.Load() {
		return
	}
	out := []any{"[DBG]"}
	if prefix := strings.Repeat("--", max(level, 0)); len(prefix) > 0 {
		out = append(out, prefix)
	}
	out = append(out, "<"+strings.ToUpper(message)+">")
	out = append(out, args...)
	fmt.Fprintln(os.Stdout, out...) //nolint:errcheck
}
