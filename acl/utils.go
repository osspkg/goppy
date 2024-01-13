/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package acl

import (
	"strconv"
	"strings"
)

const MaxLevel = uint8(9)

func validateLevel(v uint8) uint8 {
	if v > MaxLevel {
		return MaxLevel
	}
	return v
}

func str2uint(data string) []uint8 {
	t := make([]uint8, len(data))
	for i, s := range strings.Split(data, "") {
		v, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			t[i] = 0
			continue
		}
		b := uint8(v)
		if b > MaxLevel {
			t[i] = 9
		} else {
			t[i] = uint8(b)
		}
	}
	return t
}

func uint2str(data ...uint8) string {
	t := ""
	for _, v := range data {
		if v > MaxLevel {
			v = MaxLevel
		}
		t += strconv.FormatUint(uint64(v), 10)
	}
	return t
}
