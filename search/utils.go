/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import "regexp"

var (
	rexIndexName      = regexp.MustCompile(`^[a-z0-9\_]+$`)
	rexUpperCamelCase = regexp.MustCompile(`^[A-Z][A-Za-z0-9]+$`)
)

func isValidIndexName(v string) bool {
	return rexIndexName.MatchString(v)
}

func isUpperCamelCase(v string) bool {
	return rexUpperCamelCase.MatchString(v)
}
