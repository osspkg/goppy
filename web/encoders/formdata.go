/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package encoders

import (
	"encoding"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/data"
)

/***********************************************************************************************************************
********************************************* FORM DATA ENCODER ********************************************************
***********************************************************************************************************************/

type (
	bytesGetter interface {
		Bytes() []byte
	}

	FormData struct {
		funcs       []func(w *multipart.Writer) error
		buffer      *data.Buffer
		contentType string
	}
)

func (fd *FormData) File(name, filename string, r io.Reader) {
	if fd == nil {
		panic("form data has not been initialized")
	}

	fd.funcs = append(fd.funcs, func(w *multipart.Writer) error {
		if s, ok := r.(io.Seeker); ok {
			if _, err := s.Seek(0, 0); err != nil {
				return fmt.Errorf("failed seek file `%s`: %w", name, err)
			}
		}

		fw, err := w.CreateFormFile(name, filepath.Base(filename))
		if err != nil {
			return fmt.Errorf("failed create form file `%s`: %w", name, err)
		}

		if _, err = io.Copy(fw, r); err != nil {
			return fmt.Errorf("failed copy form file `%s`: %w", name, err)
		}

		return nil
	})
}

func (fd *FormData) Field(name string, value any) {
	if fd == nil {
		panic("form data has not been initialized")
	}

	fd.funcs = append(fd.funcs, func(w *multipart.Writer) error {
		fw, err := w.CreateFormField(name)
		if err != nil {
			return fmt.Errorf("failed create form field `%s`: %w", name, err)
		}

		switch v := value.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64, complex64, complex128,
			bool:
			_, err = fmt.Fprintf(fw, "%v", v)
		case string:
			_, err = fw.Write([]byte(v))
		case []byte:
			_, err = fw.Write(v)
		case io.Reader:
			_, err = io.Copy(fw, v)
		case io.WriterTo:
			_, err = v.WriteTo(fw)
		case time.Time:
			_, err = fw.Write([]byte(v.Format(time.RFC3339)))
		case error:
			_, err = fmt.Fprintf(fw, "%v", v)
		case json.Marshaler:
			var b []byte
			if b, err = v.MarshalJSON(); err == nil {
				_, err = fw.Write(b)
			}
		case encoding.TextMarshaler:
			var b []byte
			if b, err = v.MarshalText(); err == nil {
				_, err = fw.Write(b)
			}
		case encoding.BinaryMarshaler:
			var b []byte
			if b, err = v.MarshalBinary(); err == nil {
				_, err = fw.Write(b)
			}
		case xml.Marshaler:
			err = v.MarshalXML(xml.NewEncoder(fw), xml.StartElement{})
		case fmt.Stringer:
			_, err = fw.Write([]byte(v.String()))
		case fmt.GoStringer:
			_, err = fw.Write([]byte(v.GoString()))
		case bytesGetter:
			_, err = fw.Write(v.Bytes())
		default:
			return fmt.Errorf("unsupported type `%T` for `%s`", value, name)
		}

		if err != nil {
			return fmt.Errorf("failed copy form field `%s`: %w", name, err)
		}

		return nil
	})
}

func (fd *FormData) Reader() io.Reader {
	if fd == nil {
		panic("form data has not been initialized")
	}

	if fd.buffer == nil {
		return nil
	}

	fd.buffer.Seek(0, 0) //nolint:errcheck

	return fd.buffer
}

func (fd *FormData) ContentType() string {
	if fd == nil {
		panic("form data has not been initialized")
	}

	if fd.buffer == nil {
		return ""
	}

	return fd.contentType
}

func (fd *FormData) Encode() error {
	if fd == nil {
		panic("form data has not been initialized")
	}

	if fd.buffer != nil {
		return nil
	}

	fd.buffer = data.NewBuffer(512)
	mpw := multipart.NewWriter(fd.buffer)
	defer mpw.Close() //nolint:errcheck

	for _, field := range fd.funcs {
		if err := field(mpw); err != nil {
			return err
		}
	}

	fd.contentType = mpw.FormDataContentType()

	return nil
}

/***********************************************************************************************************************
********************************************* FORM DATA DECODER ********************************************************
***********************************************************************************************************************/

func FormDataDecode(r *http.Request, maxMemory int64, obj any) error {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		return fmt.Errorf("failed parse multipart form: %w", err)
	}

	modelRef := reflect.ValueOf(obj)
	if modelRef.Kind() != reflect.Ptr || modelRef.IsNil() {
		return fmt.Errorf("got non-pointer or nil-pointer object")
	}

	if err := (&formDataParser{httpReq: r, objRef: modelRef}).Unmarshal(); err != nil {
		return fmt.Errorf("failed unmarshal form data: %w", err)
	}

	return nil
}

type formDataParser struct {
	httpReq *http.Request
	objRef  reflect.Value
}

const FormDataTag = "formData"

func (v *formDataParser) Unmarshal() (err error) {
	defer func() {
		if recValue := recover(); recValue != nil {
			err = fmt.Errorf("panic: %+v", recValue)
		}
	}()

	structRef := v.objRef.Type().Elem()
	for i := 0; i < structRef.NumField(); i++ {
		var value any
		err = nil

		field := structRef.Field(i)
		tag, ok := field.Tag.Lookup(FormDataTag)
		if !ok {
			continue
		}

		formFieldName, canOmitempty, isValid := v.parseTag(tag)
		if !isValid {
			return fmt.Errorf("invalid tag of field `%s`", field.Name)
		}

		switch fieldType := field.Type.String(); fieldType {
		case "io.Reader", "io.ReadSeeker":
			value, err = v.resolveFile(formFieldName)

		case "string":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return s, nil
			})

		case "*string":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return &s, nil
			})

		case "[]byte":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return []byte(s), nil
			})

		case "*[]byte":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				b := []byte(s)
				return &b, nil
			})

		case "int":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return strconv.Atoi(s)
			})

		case "*int":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				iv, er := strconv.Atoi(s)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})

		case "int64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return strconv.ParseInt(s, 10, 64)
			})

		case "*int64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				iv, er := strconv.ParseInt(s, 10, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})

		case "uint64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return strconv.ParseUint(s, 10, 64)
			})

		case "*uint64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				iv, er := strconv.ParseUint(s, 10, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})

		case "float64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return strconv.ParseFloat(s, 64)
			})

		case "*float64":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				iv, er := strconv.ParseFloat(s, 64)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})

		case "bool":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return strconv.ParseBool(s)
			})

		case "*bool":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				iv, er := strconv.ParseBool(s)
				if er != nil {
					return nil, er
				}
				return &iv, nil
			})

		case "time.Time":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				return time.Parse(s, time.RFC3339)
			})

		case "*time.Time":
			value, err = v.resolveField(formFieldName, func(s string) (any, error) {
				tv, er := time.Parse(s, time.RFC3339)
				if er != nil {
					return nil, er
				}
				return &tv, nil
			})

		default:
			return fmt.Errorf("unsupported type `%s` for field `%s`", fieldType, field.Name)
		}

		if err != nil {
			if canOmitempty {
				continue
			}
			return fmt.Errorf("field `%s`: %w", field.Name, v.prepareError(err))
		}

		v.objRef.Elem().FieldByName(field.Name).Set(reflect.ValueOf(value))
	}
	return nil
}

func (v *formDataParser) resolveFile(fieldName string) (any, error) {
	file, head, err := v.httpReq.FormFile(fieldName)
	if err != nil {
		return nil, err
	}

	buff := data.NewBuffer(int(head.Size))
	_, err = io.Copy(buff, file)

	return buff, errors.Wrap(err, file.Close())
}

func (v *formDataParser) resolveField(fieldName string, parseFunc func(s string) (any, error)) (any, error) {
	tagValue := v.httpReq.FormValue(fieldName)
	if len(tagValue) == 0 {
		return nil, fmt.Errorf("field `%s` not found", fieldName)
	}

	return parseFunc(tagValue)
}

func (*formDataParser) parseTag(v string) (name string, omitEmpty, isValid bool) {
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

func (*formDataParser) prepareError(err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "http: no such file":
		return fmt.Errorf("not found")
	}

	return err
}

/***********************************************************************************************************************
********************************************* FORM DATA PARSER *********************************************************
***********************************************************************************************************************/
