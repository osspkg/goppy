package util

import (
	"fmt"
	"hash/crc32"
	"strings"
	"unicode"
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

	return fmt.Sprintf("%08X", crc32.Checksum([]byte(s), crc32q))
}
