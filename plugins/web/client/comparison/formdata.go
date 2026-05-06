/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"io"

	"go.osspkg.com/goppy/v3/plugins/web/encoders"
)

type FORMDATA struct {
	MaxMemory int64
}

func (v FORMDATA) Encode(w io.Writer, in any) (string, error) {
	fd, ok := in.(encoders.FormDataMarshaler)
	if !ok {
		return "", NoCast
	}

	return fd.MarshalFormData(w)
}

func (v FORMDATA) Decode(r io.Reader, out any) error {
	fd, ok := out.(encoders.FormDataUnmarshaler)
	if !ok {
		return NoCast
	}

	return fd.UnmarshalFormData(r, v.MaxMemory)
}
