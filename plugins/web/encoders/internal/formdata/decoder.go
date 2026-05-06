/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package formdata

import (
	"fmt"
	"io"
	"os"
	"reflect"

	"go.osspkg.com/cast"
	"go.osspkg.com/errors"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(r io.Reader, maxmem int64, arg any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrap(err, fmt.Errorf("panic: %+v", e))
		}
	}()

	ref := reflect.ValueOf(arg)
	if ref.Kind() != reflect.Ptr || ref.IsNil() {
		err = fmt.Errorf("got non-pointer or nil-pointer object")
		return
	}

	if ref.Elem().Kind() != reflect.Struct {
		err = fmt.Errorf("got non-struct object")
		return
	}

	var file *os.File
	defer func() {
		if file != nil {
			err = errors.Wrap(err, file.Close(), os.Remove(file.Name()))
		}
	}()

	file, err = os.CreateTemp(os.TempDir(), "multipart-*")
	if err != nil {
		return
	}

	var (
		n int64
		e error
	)
	if n, e = io.CopyN(file, r, maxmem+1); e != nil && !errors.Is(e, io.EOF) {
		err = e
		return
	}
	err = nil
	if n == 0 {
		return errFormIsEmpty
	}
	if n > maxmem {
		return errFormTooLarge
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}

	var boundary string
	if boundary, err = getBoundary(file); err != nil {
		return
	}

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}

	mp := newMultipartReadAdapter(file, boundary)
	if err = mp.ParseMultipartForm(maxmem); err != nil {
		return
	}

	err = d.parse(mp, ref)
	err = errors.Wrap(err, mp.Close())

	return
}

func (d *Decoder) parse(r mpReader, ref reflect.Value) error {
	structRef := ref.Elem().Type()

	for i := 0; i < structRef.NumField(); i++ {
		field := structRef.Field(i)
		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup(TagForm)
		if !ok {
			continue
		}

		fieldName, canOmitempty, isValid := parseTag(tag)
		if !isValid {
			return fmt.Errorf("invalid tag of field `%s`", field.Name)
		}

		if err := d.apply(ref, r, fieldName, canOmitempty, field, ref.Elem().Field(i).Interface()); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) apply(w reflect.Value, r mpReader, name string, omit bool, field reflect.StructField, obj any) error {

	var (
		err error
		out any
		set bool
	)

	switch x := obj.(type) {

	case io.Writer:
		err = r.FormFile(name, func(value io.Reader, size int64) error {
			_, e := io.Copy(x, value)
			return e
		})

	case io.ReaderFrom:
		err = r.FormFile(name, func(value io.Reader, size int64) error {
			_, e := x.ReadFrom(value)
			return e
		})

	default:
		set = true
		err = r.FormValue(name, func(value string) error {
			result := reflect.New(reflect.TypeOf(obj))
			if e := cast.StringDecode(result.Interface(), value); e != nil {
				return e
			}
			out = result.Elem().Interface()
			return nil
		})
	}

	if err != nil {
		if omit {
			return nil
		}
		return err
	}

	if set {
		w.Elem().FieldByName(field.Name).Set(reflect.ValueOf(out))
	}

	return nil
}
