/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import (
	"fmt"
	"reflect"

	"go.osspkg.com/algorithms/graph/kahn"
	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xc"
)

type (
	container struct {
		kahn   *kahn.Graph
		srv    *serviceTree
		store  *objectStorage
		status iosync.Switch
	}

	Container interface {
		Start() error
		Register(items ...interface{}) error
		Invoke(item interface{}) error
		Stop() error
	}
)

func NewContainer(ctx xc.Context) Container {
	return &container{
		kahn:   kahn.New(),
		srv:    newServiceTree(ctx),
		store:  newObjectStorage(),
		status: iosync.NewSwitch(),
	}
}

// Stop - stop all services in dependencies
func (v *container) Stop() error {
	if !v.status.Off() {
		return nil
	}
	return v.srv.Down()
}

// Start - initialize dependencies and start
func (v *container) Start() error {
	if !v.status.On() {
		return errDepAlreadyRunned
	}
	if err := v.srv.MakeAsUp(); err != nil {
		return err
	}
	if err := v.prepare(); err != nil {
		return err
	}
	if err := v.kahn.Build(); err != nil {
		return errors.Wrapf(err, "dependency graph calculation")
	}
	return v.run()
}

func (v *container) Register(items ...interface{}) error {
	if v.srv.IsOn() {
		return errDepAlreadyRunned
	}

	for _, item := range items {
		ref := reflect.TypeOf(item)
		rt := asTypeExist
		switch ref.Kind() {
		case reflect.Func, reflect.Struct:
			rt = asTypeNew
		default:
		}
		if err := v.store.Add(ref, item, rt); err != nil {
			return err
		}
	}
	return nil
}

var root = "ROOT"

func (v *container) prepare() error {
	return v.store.Each(func(item *objectStorageItem) error {

		switch item.Kind {

		case reflect.Func:
			if item.ReflectType.NumIn() == 0 {
				v.kahn.Add(root, item.Address)
				break
			}
			for i := 0; i < item.ReflectType.NumIn(); i++ {
				inRefType := item.ReflectType.In(i)
				inAddress, _ := getReflectAddress(inRefType, nil)
				v.kahn.Add(inAddress, item.Address)
			}
			for i := 0; i < item.ReflectType.NumOut(); i++ {
				outRefType := item.ReflectType.Out(i)
				outAddress, _ := getReflectAddress(outRefType, nil)
				v.kahn.Add(item.Address, outAddress)
			}

		case reflect.Struct:
			if item.ReflectType.NumField() == 0 {
				v.kahn.Add(root, item.Address)
				break
			}
			for i := 0; i < item.ReflectType.NumField(); i++ {
				inRefType := item.ReflectType.Field(i).Type
				inAddress, _ := getReflectAddress(inRefType, nil)
				v.kahn.Add(inAddress, item.Address)
			}

		default:
			v.kahn.Add(root, item.Address)
		}

		return nil
	})
}

func (v *container) Invoke(obj interface{}) error {
	if v.srv.IsOff() {
		return errDepNotRunning
	}
	item, err := v.toStoreItem(obj)
	if err != nil {
		return err
	}
	_, err = v.callArgs(item)
	if err != nil {
		return err
	}
	if item.Service == itDownService {
		if err = v.srv.AddAndUp(item.Value); err != nil {
			return err
		}
	}
	return nil
}

func (v *container) toStoreItem(obj interface{}) (*objectStorageItem, error) {
	item, ok := obj.(*objectStorageItem)
	if ok {
		return item, nil
	}
	ref := reflect.TypeOf(obj)
	address, ok := getReflectAddress(ref, obj)
	if !ok {
		return nil, fmt.Errorf("dependency [%s] is not supported", address)
	}
	serviceStatus := itNotService
	if isService(obj) {
		serviceStatus = itDownService
	}
	item = &objectStorageItem{
		Address:      address,
		RelationType: 0,
		ReflectType:  ref,
		Kind:         ref.Kind(),
		Value:        obj,
		Service:      serviceStatus,
	}
	return item, nil
}

func (v *container) callArgs(obj interface{}) ([]reflect.Value, error) {
	item, err := v.toStoreItem(obj)
	if err != nil {
		return nil, err
	}

	switch item.Kind {

	case reflect.Func:
		args := make([]reflect.Value, 0, item.ReflectType.NumIn())
		for i := 0; i < item.ReflectType.NumIn(); i++ {
			inRefType := item.ReflectType.In(i)
			inAddress, ok := getReflectAddress(inRefType, item.Value)
			if !ok {
				return nil, fmt.Errorf("dependency [%s] is not supported", inAddress)
			}
			dep, err := v.store.Get(inAddress)
			if err != nil {
				return nil, err
			}
			args = append(args, reflect.ValueOf(dep.Value))
		}
		args = reflect.ValueOf(item.Value).Call(args)
		for _, arg := range args {
			if err, ok := arg.Interface().(error); ok && err != nil {
				return nil, err
			}
		}
		return args, nil

	case reflect.Struct:
		value := reflect.New(item.ReflectType)
		args := make([]reflect.Value, 0, 1)
		for i := 0; i < item.ReflectType.NumField(); i++ {
			inRefType := item.ReflectType.Field(i)
			inAddress, ok := getReflectAddress(inRefType.Type, nil)
			if !ok {
				return nil, fmt.Errorf("dependency [%s] is not supported", inAddress)
			}
			dep, err := v.store.Get(inAddress)
			if err != nil {
				return nil, err
			}
			value.Elem().FieldByName(inRefType.Name).Set(reflect.ValueOf(dep.Value))
		}
		return append(args, value.Elem()), nil

	default:
	}

	return []reflect.Value{reflect.ValueOf(item.Value)}, nil
}

// nolint: gocyclo
func (v *container) run() error {
	names := make(map[string]struct{})
	for _, name := range v.kahn.Result() {
		if name == root {
			continue
		}
		names[name] = struct{}{}
	}

	for _, name := range v.kahn.Result() {
		if name == root {
			continue
		}
		item, err := v.store.Get(name)
		if err != nil {
			return err
		}
		if item.RelationType == asTypeExist {
			if item.Service == itDownService {
				if err = v.srv.AddAndUp(item.Value); err != nil {
					return errors.Wrapf(err, "service initialization error [%s]", item.Address)
				}
				item.Service = itUpedService
			}
			delete(names, name)
			continue
		}
		args, err := v.callArgs(item)
		if err != nil {
			return errors.Wrapf(err, "initialize error [%s]", name)
		}
		for _, arg := range args {
			address, ok := getReflectAddress(arg.Type(), arg.Interface())
			if !ok {
				if address == "error" {
					continue
				}
				return fmt.Errorf("dependency [%s] is not supported form [%s]", address, item.Address)
			}
			if isService(arg.Interface()) {
				if err = v.srv.AddAndUp(arg.Interface()); err != nil {
					return errors.Wrapf(err, "service initialization error [%s]", address)
				}
			}
			delete(names, address)
			if err = v.store.Add(arg.Type(), arg.Interface(), asTypeExist); err != nil {
				return errors.Wrapf(err, "initialize error [%s]", address)
			}
		}
		delete(names, name)
	}

	v.srv.IterateOver()
	return nil
}
