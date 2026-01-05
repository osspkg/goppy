/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"fmt"
	"net/http"

	"go.osspkg.com/logx"
)

func StringEncode(w http.ResponseWriter, code int, obj string, args ...any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	if _, err := fmt.Fprintf(w, obj, args...); err != nil {
		logx.Error("web.encoders.StringEncode", "err", err)
	}
}
