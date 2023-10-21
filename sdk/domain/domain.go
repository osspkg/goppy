/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package domain

var dot = byte('.')

func Level(s string, level int) string {
	max := len(s) - 1
	count, pos := 0, 0
	if s[max] == dot {
		max--
	}

	for i := max; i >= 0; i-- {
		if s[i] == dot {
			count++
			if count == level {
				pos = i + 1
				break
			}
		}
	}
	return s[pos:]
}
