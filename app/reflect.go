/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"fmt"
	"reflect"
)

const errName = "error"

var errType = reflect.TypeOf(new(error)).Elem()

func getReflectAddress(t reflect.Type, v interface{}) (string, bool) {
	if len(t.PkgPath()) > 0 {
		return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name()), true
	}
	switch t.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Ptr, reflect.Slice:
		if t.Implements(errType) {
			return errName, false
		}
		if len(t.Elem().PkgPath()) > 0 {
			return fmt.Sprintf("%s.%s", t.Elem().PkgPath(), t.Elem().Name()), true
		}
	case reflect.Func:
		if v == nil {
			return t.String(), false
		}
		p := reflect.ValueOf(v).Pointer()
		return fmt.Sprintf("0x%x.%s", p, t.String()), true
	}
	return t.String(), false
}

func typingReflectPtr(vv []interface{}, call func(interface{}) error) ([]interface{}, error) {
	result := make([]interface{}, 0, len(vv))
	for _, v := range vv {
		ref := reflect.TypeOf(v)
		switch ref.Kind() {
		case reflect.Struct:
			in := reflect.New(ref).Interface()
			if err := call(in); err != nil {
				return nil, err
			}
			rv := reflect.ValueOf(in).Elem().Interface()
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
