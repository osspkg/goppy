/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"io"
)

type Type interface {
	Encode(w io.Writer, in any) (contentType string, err error)
	Decode(r io.Reader, out any) (err error)
}

var NoCast = &noCastErr{}

type noCastErr struct{}

func (noCastErr) Error() string { return "no cast type" }
