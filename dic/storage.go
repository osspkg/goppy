/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"fmt"
	"reflect"

	"go.osspkg.com/syncing"
)

type object struct {
	Address string
	Type    reflect.Type
	Value   reflect.Value
}

func objectFromAny(arg any) *object {
	obj := &object{}

	if arg == nil {
		obj.Type = nil
		obj.Value = reflect.Value{}
	} else {
		obj.Type = reflect.TypeOf(arg)
		obj.Value = reflect.ValueOf(arg)
	}

	obj.Address = ResolveAddress(obj.Type, obj.Value)

	return obj
}

type storage struct {
	data *syncing.Map[string, *object]
}

func newStorage() *storage {
	return &storage{
		data: syncing.NewMap[string, *object](10),
	}
}

func (v *storage) Yield(call func(any)) {
	for _, obj := range v.data.Yield() {
		if obj == nil || !obj.Value.IsValid() {
			continue
		}
		call(obj.Value.Interface())
	}
}

func (v *storage) Get(address string) (*object, error) {
	if item, ok := v.data.Get(address); ok {
		return item, nil
	}
	return nil, fmt.Errorf("dependency [%s] not exist", address)
}

func (v *storage) GetCollection(ref reflect.Type) reflect.Value {
	sliceType := ref.Elem()
	sliceValue := reflect.MakeSlice(reflect.SliceOf(sliceType), 0, 2)

	for _, item := range v.data.Yield() {
		if item.Value.IsValid() && item.Type.Implements(sliceType) {
			sliceValue = reflect.Append(sliceValue, item.Value)
			continue
		}

		if item.Value.Kind() == reflect.Slice {
			for i := 0; i < item.Value.Len(); i++ {
				if elm := item.Value.Index(i); elm.Elem().CanConvert(sliceType) {
					sliceValue = reflect.Append(sliceValue, elm.Elem())
				}
			}
		}
	}

	return sliceValue
}

func (v *storage) Set(arg *object) error {
	if arg == nil {
		return nil
	}

	if found, ok := v.data.Get(arg.Address); ok {
		switch {
		case found.Value.IsValid() && !arg.Value.IsValid() ||
			!found.Value.IsValid() && !arg.Value.IsValid():
			//nothing do
			return nil
		case !found.Value.IsValid() && arg.Value.IsValid():
			// can replace
		default:
			return fmt.Errorf("dependency [%s] already initiated", arg.Address)
		}
	}

	v.data.Set(arg.Address, arg)

	return nil
}
