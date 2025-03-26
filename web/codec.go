/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils"
)

func JSONEncode(w http.ResponseWriter, obj any) {
	b, err := json.Marshal(obj)
	if err != nil {
		ErrorEncode(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b) // nolint: errcheck
}

func JSONDecode(r *http.Request, obj any) error {
	b, err := ioutils.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)
}

func XMLEncode(w http.ResponseWriter, obj any) {
	b, err := xml.Marshal(obj)
	if err != nil {
		ErrorEncode(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(b) // nolint: errcheck
}

func XMLDecode(r *http.Request, obj any) error {
	b, err := ioutils.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, obj)
}

func ErrorEncode(w http.ResponseWriter, obj error) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(obj.Error())) // nolint: errcheck
}

func StreamEncode(w http.ResponseWriter, obj []byte, filename string) {
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write(obj) // nolint: errcheck
}

func RawEncode(w http.ResponseWriter, obj []byte) {
	w.Header().Add("Content-Type", http.DetectContentType(obj))
	w.WriteHeader(http.StatusOK)
	w.Write(obj) // nolint: errcheck
}

func FormDataDecode(r *http.Request, obj any) error {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return fmt.Errorf("parse multipart form: %w", err)
	}
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("got non-pointer or nil-pointer object")
	}
	fd := &formData{req: r, ref: rv}
	return fd.build()
}

// ---------------------------------------------------------------------------------------------------------------------

type formData struct {
	req *http.Request
	ref reflect.Value
}

func (v *formData) build() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %+v", e)
		}
	}()

	rt := v.ref.Type().Elem()
	for i := 0; i < rt.NumField(); i++ {
		var value any
		err = nil

		field := rt.Field(i)
		tag, ok := field.Tag.Lookup("formData")
		if !ok {
			continue
		}

		tagName, omitempty, isValid := v.parseTag(tag)
		if !isValid {
			return fmt.Errorf("invalid tag of field `%s`", field.Name)
		}

		switch field.Type.String() {
		case "io.Reader":
			value, err = v.formFile(tagName)
		case "string":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return s, nil
			})
		case "*string":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return &s, nil
			})
		case "[]byte":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return []byte(s), nil
			})
		case "*[]byte":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				b := []byte(s)
				return &b, nil
			})
		case "int":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return strconv.Atoi(s)
			})
		case "*int":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				iv, er := strconv.Atoi(s)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})
		case "int64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return strconv.ParseInt(s, 10, 64)
			})
		case "*int64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				iv, er := strconv.ParseInt(s, 10, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})
		case "uint64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return strconv.ParseUint(s, 10, 64)
			})
		case "*uint64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				iv, er := strconv.ParseUint(s, 10, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})
		case "float64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return strconv.ParseFloat(s, 64)
			})
		case "*float64":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				iv, er := strconv.ParseFloat(s, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})
		case "bool":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				return strconv.ParseBool(s)
			})
		case "*bool":
			value, err = v.formValue(tagName, func(s string) (any, error) {
				iv, er := strconv.ParseBool(s)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})
		default:
			return fmt.Errorf("unsupported type `%s` for field `%s`", field.Type.String(), field.Name)
		}

		if err != nil {
			if omitempty {
				continue
			}
			return fmt.Errorf("field value `%s`: %w", field.Name, v.prepareError(err))
		}

		v.ref.Elem().FieldByName(field.Name).Set(reflect.ValueOf(value))
	}
	return nil
}

func (v *formData) formFile(tagName string) (any, error) {
	file, _, err := v.req.FormFile(tagName)
	if err != nil {
		return nil, err
	}
	buff := bytes.NewBuffer(nil)
	_, err = io.Copy(buff, file)
	return buff, errors.Wrap(err, file.Close())
}

func (v *formData) formValue(tagName string, parseFunc func(s string) (any, error)) (value any, err error) {
	tagValue := v.req.FormValue(tagName)
	if len(tagValue) == 0 {
		err = fmt.Errorf("not found")
		return
	}
	value, err = parseFunc(tagValue)
	return
}

func (*formData) parseTag(v string) (name string, omitEmpty, isValid bool) {
	isValid = true

	vs := strings.Split(v, ",")
	switch len(vs) {
	case 0:
	case 1:
		name, omitEmpty = vs[0], false
	default:
		name, omitEmpty = vs[0], strings.TrimSpace(vs[1]) == "omitempty"
	}

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		isValid = false
	}
	return
}

func (*formData) prepareError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "http: no such file":
		return fmt.Errorf("not found")
	}

	return err
}
