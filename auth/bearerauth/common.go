/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package bearerauth

import (
	"fmt"
	"strings"

	"go.osspkg.com/errors"
)

const (
	name  = `Authorization`
	value = `Bearer `
)

type (
	Writer interface {
		Set(key, value string)
	}

	Reader interface {
		Get(key string) string
	}
)

func Decode(r Reader) (string, error) {
	val := r.Get(name)
	if len(val) == 0 {
		return "", fmt.Errorf("header `%s` not found", name)
	}

	if !strings.HasPrefix(val, value) {
		return "", errors.New("invalid format bearer auth")
	}

	val = strings.TrimSpace(val[len(value):])

	return val, nil
}

func Encode(w Writer, data string) {
	w.Set(name, value+strings.TrimSpace(data))
}
