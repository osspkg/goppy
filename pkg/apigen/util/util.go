/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package util

import (
	"fmt"
	"hash/crc32"
	"strings"
	"unicode"

	"go.osspkg.com/errors"
)

func ToKebabCase(s string) string {
	if s == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(s) + 2)

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(unicode.ToLower(r))
		} else {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func ToUpperCamelCase(s string) string {
	if s == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(s) + 2)

	for i, r := range s {
		if i == 0 && !unicode.IsUpper(r) {
			builder.WriteRune(unicode.ToUpper(r))
		} else {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

var crc32q = crc32.MakeTable(0xD5828281)

func CRC32(s string) string {
	bl := len(s)

	if bl <= 0 {
		return ""
	}

	return fmt.Sprintf("%08x", crc32.Checksum([]byte(s), crc32q))
}

func SplitLast(s, sep string) string {
	result := strings.Split(s, sep)
	return result[len(result)-1]
}

func PanicIfError(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}
	err = errors.Wrapf(err, msg, args...)
	panic(err.Error())
}
