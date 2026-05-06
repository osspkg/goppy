/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package formdata

import (
	"fmt"
	"io"
	"mime/multipart"
	"reflect"

	"go.osspkg.com/cast"
)

type Encoder struct{}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func (enc *Encoder) Encode(w io.Writer, arg any) (ct string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %+v", e)
		}
	}()

	ref := reflect.ValueOf(arg)
	if ref.Kind() == reflect.Ptr {
		if ref.IsNil() {
			err = fmt.Errorf("got nil-pointer object")
			return
		}
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		err = fmt.Errorf("got non-struct object")
		return
	}

	return enc.parse(w, ref)
}

func (enc *Encoder) parse(out io.Writer, ref reflect.Value) (ct string, err error) {
	w := multipart.NewWriter(out)
	defer func() {
		ct = w.FormDataContentType()
		w.Close() //nolint:errcheck
	}()

	refType := ref.Type()

	for i := 0; i < refType.NumField(); i++ {
		field := refType.Field(i)
		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup(TagForm)
		if !ok {
			continue
		}

		fieldName, _, ok := parseTag(tag)
		if !ok {
			continue
		}

		if err = enc.apply(w, fieldName, field, ref.Field(i).Interface()); err != nil {
			return
		}
	}

	return
}

func (enc *Encoder) apply(w *multipart.Writer, name string, field reflect.StructField, value any) error {
	if value == nil {
		return nil
	}

	var (
		fw  io.Writer
		err error
	)

	switch ff := value.(type) {
	case io.Reader:
		if s, ok := value.(io.Seeker); ok {
			if _, err = s.Seek(0, 0); err != nil {
				return fmt.Errorf("failed seek file `%s`: %w", name, err)
			}
		}
		filename := "file.raw"
		if fn, ok := value.(fileNamer); ok {
			filename = fn.FileName()
		} else if fsn, _, ok := parseTag(field.Tag.Get(TagFile)); ok {
			filename = fsn
		}
		if fw, err = w.CreateFormFile(name, filename); err != nil {
			return fmt.Errorf("failed create form field `%s`: %w", name, err)
		}
		_, err = io.Copy(fw, ff)

	default:
		if fw, err = w.CreateFormField(name); err != nil {
			return fmt.Errorf("failed create form field `%s`: %w", name, err)
		}

		var s string
		if s, err = cast.StringEncode(value); err != nil {
			return fmt.Errorf("failed encode value `%s`: %w", name, err)
		}

		_, err = io.WriteString(fw, s)
	}

	if err != nil {
		return err
	}
	return nil
}
