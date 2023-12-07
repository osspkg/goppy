/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package headers

type (
	obj struct {
		data map[string][]string
	}

	Headers interface {
		String() string
	}
)

func New() Headers {
	return &obj{
		data: make(map[string][]string),
	}
}

func (v *obj) String() string {
	return ""
}