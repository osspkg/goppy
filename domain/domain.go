/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var dot = byte('.')

func Level(s string, level int) string {
	if level == 0 {
		return "."
	}
	var err error
	s, err = Normalize(s)
	if err != nil {
		return "."
	}
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

func CountLevels(s string) int {
	ss, err := Normalize(s)
	if err != nil {
		return 0
	}
	return strings.Count(ss, ".")
}

var domainRegexp = regexp.MustCompile(`^(?i)([a-z0-9-]+\.?)+$`)

func IsValid(d string) bool {
	return domainRegexp.MatchString(d)
}

func Normalize(domain string) (string, error) {
	domain = strings.TrimSpace(domain)
	if !domainRegexp.MatchString(domain) {
		return "", fmt.Errorf("invalid domain")
	}
	domain = strings.TrimRight(domain, ".")
	domain = strings.ToLower(domain)
	return domain + ".", nil
}
