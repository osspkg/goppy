/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package comparison

import (
	"encoding"
	"fmt"
	"io"
	"net/http"
)

const sniffLen = 512

type BYTES struct{}

func (BYTES) Encode(w io.Writer, in any) (string, error) {
	switch data := in.(type) {
	case []byte:
		if _, err := w.Write(data); err != nil {
			return "", err
		}
		if len(data) > sniffLen {
			data = data[:sniffLen]
		}
		return http.DetectContentType(data), nil

	case string:
		if _, err := io.WriteString(w, data); err != nil {
			return "", err
		}
		return "text/plain; charset=utf-8", nil

	case fmt.Stringer:
		if _, err := io.WriteString(w, data.String()); err != nil {
			return "", err
		}
		return "text/plain; charset=utf-8", nil

	case fmt.GoStringer:
		if _, err := io.WriteString(w, data.GoString()); err != nil {
			return "", err
		}
		return "text/plain; charset=utf-8", nil

	case io.Reader:
		mime := make([]byte, sniffLen)
		if _, err := data.Read(mime); err != nil {
			return "", err
		}
		if _, err := w.Write(mime); err != nil {
			return "", err
		}
		if _, err := io.Copy(w, data); err != nil {
			return "", err
		}
		return http.DetectContentType(mime), nil

	case encoding.TextMarshaler:
		b, err := data.MarshalText()
		if err != nil {
			return "", err
		}
		if _, err = w.Write(b); err != nil {
			return "", err
		}
		if len(b) > sniffLen {
			b = b[:sniffLen]
		}
		return http.DetectContentType(b), nil

	case encoding.BinaryMarshaler:
		b, err := data.MarshalBinary()
		if err != nil {
			return "", err
		}
		if _, err = w.Write(b); err != nil {
			return "", err
		}
		if len(b) > sniffLen {
			b = b[:sniffLen]
		}
		return http.DetectContentType(b), nil

	case bytesGetter:
		if _, err := w.Write(data.Bytes()); err != nil {
			return "", err
		}
		return http.DetectContentType(data.Bytes()), nil

	default:
		return "", NoCast
	}
}

func (BYTES) Decode(r io.Reader, out any) (err error) {
	switch data := out.(type) {
	case *[]byte:
		*data, err = io.ReadAll(r)
		return

	case *string:
		var b []byte
		if b, err = io.ReadAll(r); err != nil {
			return
		}
		*data = string(b)

	case io.Writer:
		_, err = io.Copy(data, r)

	case encoding.TextUnmarshaler:
		var b []byte
		if b, err = io.ReadAll(r); err != nil {
			return
		}
		err = data.UnmarshalText(b)

	case encoding.BinaryUnmarshaler:
		var b []byte
		if b, err = io.ReadAll(r); err != nil {
			return
		}
		err = data.UnmarshalBinary(b)

	default:
		return NoCast
	}

	return
}

type (
	bytesGetter interface {
		Bytes() []byte
	}
)
