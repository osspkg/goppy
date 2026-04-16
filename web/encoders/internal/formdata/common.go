/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package formdata

import (
	"strings"

	"go.osspkg.com/errors"
)

const (
	TagForm = "form"
	TagFile = "filename"
)

const defaultMaxMemory = 32 << 20 // 32 MB

var (
	errMissingFile  = errors.New("no such file")
	errFormIsEmpty  = errors.New("form is empty")
	errFormTooLarge = errors.New("form too large")
)

type fileNamer interface {
	FileName() string
}

func parseTag(v string) (name string, omitEmpty, isValid bool) {
	isValid = true

	vs := strings.Split(v, ",")
	switch len(vs) {
	case 0:
	case 1:
		name, omitEmpty = vs[0], false
	default:
		name, omitEmpty = vs[0], strings.TrimSpace(vs[1]) == "omitempty"
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		isValid = false
	}
	return
}
