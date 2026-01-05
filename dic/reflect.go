/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	errorName = "error"
)

var (
	errorType = reflect.TypeOf(new(error)).Elem()
)

func isError(t reflect.Type) bool {
	return t.Implements(errorType)
}

func TypingPointer(vv []any, call func(any) error) ([]any, error) {
	result := make([]any, 0, len(vv))
	for _, v := range vv {
		ref := reflect.ValueOf(v)
		switch ref.Kind() {
		case reflect.Struct:
			in := reflect.New(ref.Type())
			in.Elem().Set(ref)
			if err := call(in.Interface()); err != nil {
				return nil, err
			}
			rv := in.Elem().Interface()
			result = append(result, rv)
		case reflect.Ptr:
			if err := call(v); err != nil {
				return nil, err
			}
			result = append(result, v)
		default:
			return nil, fmt.Errorf("supported type [%T]", v)
		}
	}
	return result, nil
}

func isInterfaceCollection(ref reflect.Type) bool {
	return ref.Kind() == reflect.Slice && ref.Elem().Kind() == reflect.Interface
}

func ResolveAddress(t reflect.Type, v reflect.Value) string {
	if t == nil {
		if !v.IsValid() {
			return "nil"
		}

		t = v.Type()
	}

	if isError(t) {
		return errorName
	}

	if t.Name() != "" && t.PkgPath() != "" {
		return t.PkgPath() + "." + t.Name()
	}

	switch t.Kind() {

	case reflect.Ptr:
		return "*" + ResolveAddress(t.Elem(), reflect.Value{})

	case reflect.Slice:
		return "[]" + ResolveAddress(t.Elem(), reflect.Value{})

	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), ResolveAddress(t.Elem(), reflect.Value{}))

	case reflect.Map:
		key := ResolveAddress(t.Key(), reflect.Value{})
		val := ResolveAddress(t.Elem(), reflect.Value{})
		return fmt.Sprintf("map[%s]%s", key, val)

	case reflect.Chan:
		prefix := "chan "
		if t.ChanDir() == reflect.RecvDir {
			prefix = "<-chan "
		} else if t.ChanDir() == reflect.SendDir {
			prefix = "chan<- "
		}
		return prefix + ResolveAddress(t.Elem(), reflect.Value{})

	case reflect.Func:
		if !v.IsValid() {
			return buildFuncSignature(t)
		}
		return fmt.Sprintf("0x%x.%s", v.Pointer(), buildFuncSignature(t))

	case reflect.Struct:
		if t.NumField() == 0 {
			return "struct{}"
		}
		return t.String()

	default:
		return t.String()
	}
}

func buildFuncSignature(t reflect.Type) string {
	var b strings.Builder
	b.WriteString("func(")
	for i := 0; i < t.NumIn(); i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(ResolveAddress(t.In(i), reflect.Value{}))
	}
	b.WriteString(")")

	if t.NumOut() > 0 {
		b.WriteString(" ")
		if t.NumOut() == 1 {
			b.WriteString(ResolveAddress(t.Out(0), reflect.Value{}))
		} else {
			b.WriteString("(")
			for i := 0; i < t.NumOut(); i++ {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(ResolveAddress(t.Out(i), reflect.Value{}))
			}
			b.WriteString(")")
		}
	}
	return b.String()
}
