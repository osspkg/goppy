/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package common

import (
	"fmt"
	"io"
	"strings"

	"go.osspkg.com/console"
)

func SplitLast(s, sep string) string {
	result := strings.Split(s, sep)
	return result[len(result)-1]
}

func Writef(w io.Writer, s string, args ...any) {
	_, err := fmt.Fprintf(w, s, args...)
	console.FatalIfErr(err, "orm builder")
}

func Write(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	console.FatalIfErr(err, "orm builder")
}

func Writelnf(w io.Writer, s string, args ...any) {
	_, err := fmt.Fprintf(w, s+"\n", args...)
	console.FatalIfErr(err, "orm builder")
}

func Writeln(w io.Writer, s string) {
	_, err := io.WriteString(w, s+"\n")
	console.FatalIfErr(err, "orm builder")
}
