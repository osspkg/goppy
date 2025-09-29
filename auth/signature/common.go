/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature

import (
	"fmt"
	"regexp"
)

const (
	headerName  = `Signature`
	valueRegexp = `keyId=\"(.*)\",algorithm=\"(.*)\",signature=\"(.*)\"`
	valueTmpl   = `keyId="%s",algorithm="%s",signature="%s"`
)

var (
	rex = regexp.MustCompile(valueRegexp)
)

type (
	Data struct {
		ID  string
		Alg string
		Sig string
	}

	Writer interface {
		Set(key, value string)
	}

	Reader interface {
		Get(key string) string
	}
)

// Decode getting signature from header
func Decode(r Reader) (*Data, error) {
	val := r.Get(headerName)
	if len(val) == 0 {
		return nil, fmt.Errorf("header `%s` not found", headerName)
	}

	submatch := rex.FindSubmatch([]byte(val))
	if len(submatch) != 4 {
		return nil, fmt.Errorf("invalid format header `%s`", headerName)
	}

	return &Data{
		ID:  string(submatch[1]),
		Alg: string(submatch[2]),
		Sig: string(submatch[3]),
	}, nil
}

// Encode make and setting signature to header
func Encode(w Writer, s Signature, body []byte) error {
	sig, err := s.Create(body)
	if err != nil {
		return fmt.Errorf("create signature: %w", err)
	}

	w.Set(headerName, fmt.Sprintf(valueTmpl, s.ID(), s.Algorithm(), sig))
	return nil
}
