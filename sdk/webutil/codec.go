/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package webutil

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"github.com/osspkg/goppy/sdk/ioutil"
)

func JSONEncode(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		ErrorEncode(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b) //nolint: errcheck
}

func JSONDecode(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func XMLEncode(w http.ResponseWriter, v interface{}) {
	b, err := xml.Marshal(v)
	if err != nil {
		ErrorEncode(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b) //nolint: errcheck
}

func XMLDecode(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, v)
}

func ErrorEncode(w http.ResponseWriter, v error) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(v.Error())) //nolint: errcheck
}

func StreamEncode(w http.ResponseWriter, v []byte, filename string) {
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write(v) //nolint: errcheck
}

func RawEncode(w http.ResponseWriter, v []byte) {
	w.Header().Add("Content-Type", http.DetectContentType(v))
	w.WriteHeader(http.StatusOK)
	w.Write(v) //nolint: errcheck
}
