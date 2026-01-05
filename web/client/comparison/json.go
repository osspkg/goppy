/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"encoding/json"
	"io"
)

const ContentTypeJSON = "application/json; charset=utf-8"

type JSON struct {
	Force bool
}

func (v JSON) Encode(w io.Writer, in any) (string, error) {
	if !v.Force {
		if _, ok := in.(json.Marshaler); !ok {
			return "", NoCast
		}
	}

	if err := json.NewEncoder(w).Encode(in); err != nil {
		return "", err
	}

	return ContentTypeJSON, nil
}

func (v JSON) Decode(r io.Reader, out any) error {
	if !v.Force {
		if _, ok := out.(json.Unmarshaler); !ok {
			return NoCast
		}
	}

	return json.NewDecoder(r).Decode(out)
}
