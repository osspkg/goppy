/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"io"
	"net/http"

	"go.osspkg.com/logx"
)

func ErrorEncode(w http.ResponseWriter, code int, obj error) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	str := ""
	if obj != nil {
		str = obj.Error()
	}

	if _, err := io.WriteString(w, str); err != nil {
		logx.Error("web.encoders.ErrorEncode", "err", err)
	}
}
