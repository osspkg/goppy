/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package base64url

import (
	"encoding/base64"
)

var b64Url = base64.URLEncoding.WithPadding(base64.NoPadding)

func Encode(b []byte) []byte {
	return b64Url.AppendEncode(nil, b)
}

func Decode(b []byte) []byte {
	dec, err := b64Url.AppendDecode(nil, b)
	if err != nil {
		return nil
	}
	return dec
}
