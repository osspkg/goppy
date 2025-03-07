/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package common

import (
	"fmt"
	"io"
	"strings"
)

func SplitLast(s, sep string) string {
	result := strings.Split(s, sep)
	return result[len(result)-1]
}

func Write(w io.Writer, s string, args ...any) {
	if _, err := fmt.Fprintf(w, s, args...); err != nil {
		panic(err)
	}
}
