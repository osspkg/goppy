/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import (
	"strings"

	"go.osspkg.com/do"
)

func SplitFirstString(div string, oneOf ...string) []string {
	var use string
	for _, one := range oneOf {
		if len(one) > 0 {
			use = one
			break
		}
	}

	if len(use) == 0 {
		return nil
	}

	return do.TreatValue(strings.Split(use, div), strings.ToLower, strings.TrimSpace)
}
