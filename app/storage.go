/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"fmt"
	"reflect"
	"sync"
)

type (
	objectRelationType    uint
	objectRelationService uint
)

const (
	asTypeNew   objectRelationType = 1
	asTypeExist objectRelationType = 2

	itNotService  objectRelationService = 0
	itDownService objectRelationService = 1
	itUpedService objectRelationService = 2
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	objectStorageItem struct {
		Address      string
		RelationType objectRelationType
		ReflectType  reflect.Type
		Kind         reflect.Kind
		Value        interface{}
		Service      objectRelationService
	}
	objectStorage struct {
		data map[string]*objectStorageItem
		mux  sync.RWMutex
	}
)

func newObjectStorage() *objectStorage {
	return &objectStorage{
		data: make(map[string]*objectStorageItem),
	}
}

func (v *objectStorage) GetByAddress(address string) (*objectStorageItem, error) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	if item, ok := v.data[address]; ok {
		return item, nil
	}
	return nil, fmt.Errorf("dependency [%s] not initiated", address)
}

func (v *objectStorage) GetByReflect(ref reflect.Type, obj interface{}) (*objectStorageItem, error) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	address, ok := getReflectAddress(ref, obj)
	if !ok {
		if address == "error" {
			return nil, errIsTypeError
		}
		return nil, fmt.Errorf("dependency [%s] is not supported", address)
	}
	return v.GetByAddress(address)
}

func (v *objectStorage) Add(ref reflect.Type, obj interface{}, relationType objectRelationType) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	address, ok := getReflectAddress(ref, obj)
	if !ok {
		if address != "error" {
			return fmt.Errorf("dependency [%s] is not supported", address)
		}
		return nil
	}
	if item, ok := v.data[address]; ok {
		if item.RelationType == asTypeExist {
			return fmt.Errorf("dependency [%s] already initiated", address)
		}
	}
	serviceStatus := itNotService
	if isService(obj) {
		serviceStatus = itDownService
	}
	v.data[address] = &objectStorageItem{
		Address:      address,
		Value:        obj,
		ReflectType:  ref,
		RelationType: relationType,
		Kind:         ref.Kind(),
		Service:      serviceStatus,
	}
	return nil
}

func (v *objectStorage) Each(call func(item *objectStorageItem) error) error {
	v.mux.RLock()
	defer v.mux.RUnlock()

	for _, item := range v.data {
		if err := call(item); err != nil {
			return err
		}
	}
	return nil
}
