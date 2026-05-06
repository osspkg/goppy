/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package basicauth

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"go.osspkg.com/errors"
)

const (
	name  = `Authorization`
	value = `Basic `
)

type (
	Writer interface {
		Set(key, value string)
	}

	Reader interface {
		Get(key string) string
	}
)

func Decode(r Reader) (string, string, error) {
	val := r.Get(name)
	if len(val) == 0 {
		return "", "", fmt.Errorf("header `%s` not found", name)
	}

	if !strings.HasPrefix(val, value) {
		return "", "", errors.New("invalid format basic auth")
	}

	val = strings.TrimSpace(val[len(value):])

	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return "", "", fmt.Errorf("faild decode basic auth: %w", err)
	}

	data := bytes.Split(b, []byte(":"))
	if len(data) != 2 {
		return "", "", errors.New("invalid format basic auth")
	}

	return string(data[0]), string(data[1]), nil
}

func Encode(w Writer, login, password string) {
	s := base64.StdEncoding.EncodeToString([]byte(login + ":" + password))

	w.Set(name, value+s)
}
