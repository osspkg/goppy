/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import "strings"

func NeedSkipFile(filepath string) bool {
	return strings.HasSuffix(filepath, "_gen.go") ||
		strings.HasSuffix(filepath, "_test.go") ||
		strings.HasSuffix(filepath, "_easyjson.go") ||
		strings.HasSuffix(filepath, "_mock.go") ||
		strings.Contains(filepath, "/vendor/")
}
