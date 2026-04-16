/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package signature

import (
	"fmt"
	"regexp"

	"go.osspkg.com/random"
)

const (
	name  = `Signature`
	value = `id=\"(.*)\" alg=\"(.*)\" sig=\"(.*)\",nonce=\"(.*)\"`
	tmpl  = `id="%s" alg="%s" sig="%s" nonce="%s"`
)

var (
	rex = regexp.MustCompile(value)
)

type (
	Data struct {
		ID    string
		Alg   string
		Sig   string
		Nonce string
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
	val := r.Get(name)
	if len(val) == 0 {
		return nil, fmt.Errorf("header `%s` not found", name)
	}

	submatch := rex.FindSubmatch([]byte(val))
	if len(submatch) != 5 {
		return nil, fmt.Errorf("invalid format header `%s`", name)
	}

	return &Data{
		ID:    string(submatch[1]),
		Alg:   string(submatch[2]),
		Sig:   string(submatch[3]),
		Nonce: string(submatch[4]),
	}, nil
}

// Encode make and setting signature to header
func Encode(w Writer, s Signature, body []byte) error {
	nonce := random.CryptoBase64(32)
	sig, err := s.Create(body, nonce)
	if err != nil {
		return fmt.Errorf("create signature: %w", err)
	}

	w.Set(name, fmt.Sprintf(tmpl, s.ID(), s.Algorithm(), sig, nonce))
	return nil
}
