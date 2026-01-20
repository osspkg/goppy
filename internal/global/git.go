/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import (
	"fmt"
	"regexp"
	"strings"
)

var rexHEAD = regexp.MustCompile(`(?mU)ref\: refs/heads/(\w+)\s+HEAD`)

func GitHEAD(url string) (string, error) {
	if len(url) == 0 {
		b, err := Exec("git remote get-url origin")
		if err != nil {
			return "", err
		}
		url = strings.Trim(string(b), "\n")
	}
	b, err := Exec("git ls-remote --symref " + url + " HEAD")
	if err != nil {
		return "", err
	}
	_strs := rexHEAD.FindStringSubmatch(string(b))
	if len(_strs) != 2 {
		return "", fmt.Errorf("HEAD not found")
	}
	return _strs[1], nil
}
