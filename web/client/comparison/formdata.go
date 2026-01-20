/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"io"

	"go.osspkg.com/goppy/v3/web/encoders"
)

type FORMDATA struct{}

func (v FORMDATA) Encode(w io.Writer, in any) (string, error) {
	fd, ok := in.(*encoders.FormData)
	if !ok {
		return "", NoCast
	}

	if err := fd.Encode(); err != nil {
		return "", err
	}

	if _, err := io.Copy(w, fd.Reader()); err != nil {
		return "", err
	}

	return fd.ContentType(), nil
}

func (v FORMDATA) Decode(r io.Reader, out any) error {
	return NoCast
}
