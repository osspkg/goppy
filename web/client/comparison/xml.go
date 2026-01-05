/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"encoding/xml"
	"io"
)

const ContentTypeXML = "application/xml; charset=utf-8"

type XML struct {
	Force        bool
	StartElement *xml.StartElement
}

func (v XML) Encode(w io.Writer, in any) (string, error) {
	if !v.Force {
		if _, ok := in.(xml.Marshaler); !ok {
			return "", NoCast
		}
	}

	enc := xml.NewEncoder(w)
	defer enc.Close() //nolint:errcheck

	if err := enc.EncodeElement(in, *v.StartElement); err != nil {
		return "", err
	}

	return ContentTypeJSON, nil
}

func (v XML) Decode(r io.Reader, out any) error {
	if !v.Force {
		if _, ok := out.(xml.Unmarshaler); !ok {
			return NoCast
		}
	}

	return xml.NewDecoder(r).DecodeElement(out, v.StartElement)
}
