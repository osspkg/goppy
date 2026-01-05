/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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

var (
	writeType = reflect.TypeOf((*io.Writer)(nil)).Elem()
	readType  = reflect.TypeOf((*io.Reader)(nil)).Elem()
)

//nolint:gocyclo
func (v *formDataParser) Unmarshal() (err error) {
	defer func() {
		if recValue := recover(); recValue != nil {
			err = fmt.Errorf("panic: %+v", recValue)
		}
	}()

	structRef := v.objRef.Type().Elem()
	for i := 0; i < structRef.NumField(); i++ {
		err = nil

		field := structRef.Field(i)
		if !field.IsExported() {
			continue
		}

		tag, ok := field.Tag.Lookup(FormDataTag)
		if !ok {
			continue
		}

		formFieldName, canOmitempty, isValid := v.parseTag(tag)
		if !isValid {
			return fmt.Errorf("invalid tag of field `%s`", field.Name)
		}

		kindOrig := field.Type.Kind()
		kindElem := field.Type.Kind()
		fieldType := field.Type.String()

		isPtr := kindOrig == reflect.Ptr

		if isPtr {
			fieldType = field.Type.Elem().String()
			kindElem = field.Type.Elem().Kind()
		}

		switch kindElem {
		case reflect.Struct:
			value := v.objRef.Elem().FieldByName(field.Name)
			isZero := value.IsZero()

			if isZero {
				if isPtr {
					value = reflect.New(v.objRef.Elem().FieldByName(field.Name).Type().Elem())
				} else {
					value = reflect.New(v.objRef.Elem().FieldByName(field.Name).Type())
				}
			}

			switch x := value.Interface().(type) {
			case io.Writer:
				err = v.resolveFile(formFieldName, func(r io.Reader, _ int) error {
					_, e := io.Copy(x, r)
					return e
				})
			case json.Unmarshaler:
				err = v.resolveField(formFieldName, func(s string) error {
					return x.UnmarshalJSON([]byte(s))
				})
			case xml.Unmarshaler:
				err = v.resolveField(formFieldName, func(s string) error {
					return xml.Unmarshal([]byte(s), x)
				})
			case encoding.TextUnmarshaler:
				err = v.resolveField(formFieldName, func(s string) error {
					return x.UnmarshalText([]byte(s))
				})
			case encoding.BinaryUnmarshaler:
				err = v.resolveField(formFieldName, func(s string) error {
					return x.UnmarshalBinary([]byte(s))
				})

			default:
				return fmt.Errorf("unsupported struct `%s` for field `%s`", fieldType, field.Name)
			}

			if err != nil {
				if canOmitempty {
					continue
				}
				return fmt.Errorf("field `%s`: %w", field.Name, v.prepareError(err))
			}

			if isZero {
				if isPtr {
					v.objRef.Elem().FieldByName(field.Name).Set(value)
				} else {
					v.objRef.Elem().FieldByName(field.Name).Set(value.Elem())
				}
			}

		case reflect.Interface:
			value := v.objRef.Elem().FieldByName(field.Name)
			isZero := value.IsZero()
			refType := value.Type()
			if !isZero {
				refType = value.Elem().Type()
			}

			switch {
			case refType.AssignableTo(writeType):
				err = v.resolveFile(formFieldName, func(r io.Reader, _ int) error {
					cp := reflect.ValueOf(io.Copy)
					args := []reflect.Value{value.Elem(), reflect.ValueOf(r)}
					out := cp.Call(args)
					var e error
					if !out[1].IsNil() {
						e = out[1].Interface().(error)
					}
					return e
				})

			case refType.AssignableTo(readType):
				err = v.resolveFile(formFieldName, func(r io.Reader, size int) error {
					w := data.NewBuffer(size)
					if _, e := io.Copy(w, r); e != nil {
						return e
					}
					v.objRef.Elem().FieldByName(field.Name).Set(reflect.ValueOf(w))
					return nil
				})

			default:
				return fmt.Errorf("unsupported interface `%s` for field `%s`", fieldType, field.Name)
			}

			if err != nil {
				if canOmitempty {
					continue
				}
				return fmt.Errorf("field `%s`: %w", field.Name, v.prepareError(err))
			}

		default:
			var value any

			switch fieldType {
			case "string":
				err = v.resolveField(formFieldName, func(s string) error {
					if isPtr {
						value = &s
					} else {
						value = s
					}
					return nil
				})

			case "[]byte":
				err = v.resolveField(formFieldName, func(s string) error {
					b := []byte(s)
					if isPtr {
						value = &b
					} else {
						value = b
					}
					return nil
				})

			case "int":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := strconv.Atoi(s)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
				})

			case "int64":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := strconv.ParseInt(s, 10, 64)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
				})
			case "uint64":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := strconv.ParseUint(s, 10, 64)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
				})

			case "float64":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := strconv.ParseFloat(s, 64)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
				})

			case "bool":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := strconv.ParseBool(s)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
				})

			case "time.Time":
				err = v.resolveField(formFieldName, func(s string) error {
					iv, e := time.Parse(s, time.RFC3339)
					if e != nil {
						return e
					}
					if isPtr {
						value = &iv
					} else {
						value = iv
					}
					return nil
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
	}
	return nil
}

func (v *formDataParser) resolveFile(fieldName string, call func(io.Reader, int) error) error {
	file, head, err := v.httpReq.FormFile(fieldName)
	if err != nil {
		return err
	}
	err = call(file, int(head.Size))
	return errors.Wrap(err, file.Close())
}

func (v *formDataParser) resolveField(fieldName string, parseFunc func(s string) error) error {
	value := v.httpReq.FormValue(fieldName)
	if len(value) == 0 {
		return fmt.Errorf("field `%s` not found", fieldName)
	}

	return parseFunc(value)
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
