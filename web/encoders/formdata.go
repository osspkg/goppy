/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"io"

	"go.osspkg.com/goppy/v3/web/encoders/internal/formdata"
)

type FormDataUnmarshaler interface {
	UnmarshalFormData(r io.Reader, maxmem int64) error
}

type FormDataMarshaler interface {
	MarshalFormData(w io.Writer) (contentType string, err error)
}

var (
	fde = formdata.NewEncoder()
	fdd = formdata.NewDecoder()
)

func FormDataEncode(w io.Writer, arg any) (string, error) {
	return fde.Encode(w, arg)
}

func FormDataDecode(r io.Reader, maxmem int64, arg any) error {
	return fdd.Decode(r, maxmem, arg)
}
