/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"encoding/xml"
	"net/http"

	"go.osspkg.com/ioutils"
	"go.osspkg.com/logx"
)

func XMLEncode(w http.ResponseWriter, code int, obj any) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(code)

	b, err := xml.Marshal(obj)
	if err != nil {
		logx.Error("web.encoders.XMLEncode", "err", err)
		return
	}
	if _, err = w.Write(b); err != nil {
		logx.Error("web.encoders.XMLEncode", "err", err)
	}
}

func XMLDecode(r *http.Request, obj any) error {
	b, err := ioutils.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, obj)
}
