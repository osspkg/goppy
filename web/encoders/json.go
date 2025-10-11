/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"encoding/json"
	"net/http"

	"go.osspkg.com/ioutils"
	"go.osspkg.com/logx"
)

func JSONEncode(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	b, err := json.Marshal(obj)
	if err != nil {
		logx.Error("web.encoders.JSONEncode", "err", err)
		return
	}
	if _, err = w.Write(b); err != nil {
		logx.Error("web.encoders.JSONEncode", "err", err)
	}
}

func JSONDecode(r *http.Request, obj json.Unmarshaler) error {
	b, err := ioutils.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}
