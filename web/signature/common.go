/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature

import (
	"fmt"
	"net/http"
	"regexp"

	"go.osspkg.com/errors"
)

const (
	Header      = `Signature`
	valueRegexp = `keyId=\"(.*)\",algorithm=\"(.*)\",signature=\"(.*)\"`
	valueTmpl   = `keyId="%s",algorithm="%s",signature="%s"`
)

var (
	ErrInvalidSignature = errors.New("invalid signature header")
	rex                 = regexp.MustCompile(valueRegexp)
)

type Data struct {
	ID   string
	Alg  string
	Hash string
}

// Decode getting signature from header
func Decode(h http.Header) (s Data, err error) {
	d := h.Get(Header)
	r := rex.FindSubmatch([]byte(d))
	if len(r) != 4 {
		err = ErrInvalidSignature
		return
	}
	s.ID, s.Alg, s.Hash = string(r[1]), string(r[2]), string(r[3])
	return
}

// Encode make and setting signature to header
func Encode(h http.Header, s Signature, body []byte) {
	h.Set(Header, fmt.Sprintf(valueTmpl, s.ID(), s.Algorithm(), s.CreateString(body)))
}
