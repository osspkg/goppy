/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package version

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

const (
	Header      = `Accept`
	valueRegexp = `application\/vnd.v(\d+)\+json`
	valueTmpl   = `application/vnd.v%d+json`
)

var rex = regexp.MustCompile(valueRegexp)

// Decode getting version from header
func Decode(h http.Header) uint64 {
	d := h.Get(Header)
	r := rex.FindSubmatch([]byte(d))
	if len(r) == 2 {
		if v, err := strconv.ParseUint(string(r[1]), 10, 64); err == nil {
			return v
		}
	}
	return 0
}

// Encode setting version to header
func Encode(h http.Header, v uint64) {
	h.Set(Header, fmt.Sprintf(valueTmpl, v))
}
