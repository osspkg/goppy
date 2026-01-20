/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"fmt"
	"io"
	"net/http"

	"go.osspkg.com/logx"
	"go.osspkg.com/static"
)

func StreamEncode(w http.ResponseWriter, code int, obj []byte, filename string) {
	w.Header().Set("Content-Type", static.DetectContentType(filename, obj))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(code)

	if _, err := w.Write(obj); err != nil {
		logx.Error("web.encoders.StreamEncode", "err", err)
	}
}

func BytesEncode(w http.ResponseWriter, code int, obj []byte) {
	w.Header().Set("Content-Type", http.DetectContentType(obj))
	w.WriteHeader(code)

	if _, err := w.Write(obj); err != nil {
		logx.Error("web.encoders.BytesEncode", "err", err)
	}
}

func ReaderEncode(w http.ResponseWriter, code int, obj io.Reader, filename string) {
	mime := make([]byte, 512)
	n, err := obj.Read(mime)
	if err != nil {
		logx.Error("web.encoders.ReaderEncode", "err", err)
	}

	if len(filename) > 0 {
		w.Header().Set("Content-Type", static.DetectContentType(filename, mime))
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	} else {
		w.Header().Set("Content-Type", http.DetectContentType(mime))
	}
	w.WriteHeader(code)

	if _, err = w.Write(mime); err != nil {
		logx.Error("web.encoders.BytesEncode", "err", err)
	}

	if n < 512 {
		return
	}

	if _, err = io.Copy(w, obj); err != nil {
		logx.Error("web.encoders.BytesEncode", "err", err)
	}
}
