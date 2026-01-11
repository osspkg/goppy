/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"

	"go.osspkg.com/algorithms/graph/kahn"
	"go.osspkg.com/errors"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/plugins"
)

const root = "ROOT"

type Container struct {
	status  syncing.Switch
	brokers []plugins.Broker
	bIndex  int
	storage *storage
	graph   *kahn.Graph
}

func New() *Container {
	return &Container{
		graph:   kahn.New(),
		brokers: make([]plugins.Broker, 0, 10),
		storage: newStorage(),
		status:  syncing.NewSwitch(),
	}
}

func (v *Container) BrokerRegister(args ...plugins.Broker) error {
	if v.status.IsOn() {
		return ErrDepAlreadyRunning
	}

	for _, arg := range args {
		for i := 0; i < len(v.brokers); i++ {
			if v.brokers[i].Name() == arg.Name() {
				return errors.Wrapf(ErrBrokerExist, "%s", arg.Name())
			}
		}
		v.brokers = append(v.brokers, arg)
	}

	slices.SortFunc(v.brokers, func(a, b plugins.Broker) int {
		return cmp.Compare(a.Priority(), b.Priority())
	})
	return nil
}

func (v *Container) Register(args ...any) error {
	if v.status.IsOn() {
		return ErrDepAlreadyRunning
	}

	for _, item := range args {
		if err := v.append(item); err != nil {
			return err
		}
	}

	return nil
}

func (v *Container) BreakPoint(arg any) error {
	if v.status.IsOn() {
		return ErrDepAlreadyRunning
	}

	obj := objectFromAny(arg)

	if obj.Type.Kind() != reflect.Func {
		return ErrBreakPointType
	}

	v.graph.BreakPoint(obj.Address)

	return nil
}

func (v *Container) Invoke(arg any) error {
	if v.status.IsOff() {
		return ErrDepNotRunning
	}

	obj := objectFromAny(arg)
	dbg(0, "Invoke", obj.Address)

	if !obj.Value.IsValid() {
		dbg(1, "validate", "got nil object value")
		return fmt.Errorf("got nil object value")
	}

	if obj.Type.Kind() != reflect.Func {
		dbg(1, "err", "is not a function")
		return ErrInvokeType
	}

	dbg(1, "Func", "in", obj.Type.NumIn(), "out", obj.Type.NumOut())

	args := make([]reflect.Value, 0, obj.Type.NumIn())
	for i := 0; i < obj.Type.NumIn(); i++ {
		inRefType := obj.Type.In(i)
		inAddress := ResolveAddress(inRefType, reflect.Value{})

		if inRefType.Kind() == reflect.Slice && isInterfaceCollection(inRefType) {
			dbg(2, "build", "arg", i, "interface collection", inAddress)
			dep := v.storage.GetCollection(inRefType)
			args = append(args, dep)
		} else {
			dbg(2, "get", "arg", i, inAddress)
			dep, err := v.storage.Get(inAddress)
			if err != nil {
				dbg(3, "err", err)
				return err
			}
			if !dep.Value.IsValid() {
				dbg(3, "validate", "got nil object value")
				return fmt.Errorf("dependency [%s] not initialized", inAddress)
			}
			args = append(args, dep.Value)
		}
	}

	dbg(2, "call")
	args = obj.Value.Call(args)

	dbg(3, "check out")
	for _, arg := range args {
		if isError(arg.Type()) {
			if err, ok := arg.Interface().(error); ok && err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Container) Start(ctx xc.Context) error {
	if !v.status.On() {
		return ErrDepAlreadyRunning
	}

	dbg(0, "Start")

	if err := v.graph.Build(); err != nil {
		dbg(1, "Build graph", "err", err)
		return errors.Wrapf(err, "dependency graph calculation")
	}

	if err := v.creatingObjects(); err != nil {
		dbg(1, "creating Objects", "err", err)
		return errors.Wrapf(err, "create objects")
	}

	for i := 0; i < len(v.brokers); i++ {
		h := v.brokers[i]

		v.storage.Yield(h.Apply)

		dbg(1, "run connector", h.Name())
		if err := h.OnStart(ctx); err != nil {
			dbg(2, "err", err)
			return errors.Wrapf(err, "run handler on start")
		}

		v.bIndex = i
	}

	return nil
}

func (v *Container) Stop() error {
	if !v.status.Off() {
		return nil
	}

	dbg(0, "Stop")

	for ; v.bIndex >= 0; v.bIndex-- {
		h := v.brokers[v.bIndex]

		if err := h.OnStop(); err != nil {
			dbg(1, "brokers", "err", err)
			return errors.Wrapf(err, "run handler on stop")
		}
	}

	return nil
}

func (v *Container) append(arg any) error {
	obj := objectFromAny(arg)
	dbg(0, "Register", obj.Address)

	if err := v.storage.Set(obj); err != nil {
		dbg(1, "Register", "err", err)
		return err
	}

	switch obj.Type.Kind() {

	case reflect.Func:

		if obj.Type.NumIn() == 0 {
			v.graph.Add(root, obj.Address)
			dbg(1, "graph", "func", root, "->", obj.Address)
		}
		for i := 0; i < obj.Type.NumIn(); i++ {
			inRefType := obj.Type.In(i)
			inAddress := ResolveAddress(inRefType, reflect.Value{})
			v.graph.Add(inAddress, obj.Address)
			dbg(1, "graph", "func-in", inAddress, "->", obj.Address)
			if err := v.storage.Set(&object{Address: inAddress, Type: inRefType}); err != nil {
				dbg(2, "graph", "err", err)
				return err
			}
		}

		for i := 0; i < obj.Type.NumOut(); i++ {
			outRefType := obj.Type.Out(i)
			outAddress := ResolveAddress(outRefType, reflect.Value{})
			v.graph.Add(obj.Address, outAddress)
			dbg(1, "graph", "func-out", obj.Address, "->", outAddress)
			if err := v.storage.Set(&object{Address: outAddress, Type: outRefType}); err != nil {
				dbg(2, "graph", "err", err)
				return err
			}
		}

	case reflect.Struct:
		obj.Value = reflect.Value{}

		if obj.Type.NumField() == 0 {
			v.graph.Add(root, obj.Address)
			dbg(1, "graph", "struct", root, "->", obj.Address)
		}
		for i := 0; i < obj.Type.NumField(); i++ {
			inRefType := obj.Type.Field(i).Type
			inAddress := ResolveAddress(inRefType, reflect.Value{})
			v.graph.Add(inAddress, obj.Address)
			dbg(1, "graph", "strict", inAddress, "->", obj.Address)
			if err := v.storage.Set(&object{Address: inAddress, Type: inRefType}); err != nil {
				dbg(2, "graph", "err", err)
				return err
			}
		}

	default:
		v.graph.Add(root, obj.Address)
		dbg(1, "graph", "any", root, "->", obj.Address)
	}

	return nil
}

func (v *Container) creatingObjects() error {
	for _, objectName := range v.graph.Result() {

		switch objectName {
		case root, errorName:
			continue
		default:
		}

		dbg(0, "Creating Objects", objectName)

		obj, err := v.storage.Get(objectName)
		if err != nil {
			dbg(1, "err", err)
			return errors.Wrapf(err, "object [%s] not found", objectName)
		}

		dbg(1, "initialize", objectName)

		if err = v.initializeObject(obj); err != nil {
			dbg(2, "err", err)
			return errors.Wrapf(err, "failed initialize object [%s]", objectName)
		}
	}

	return nil
}

func (v *Container) initializeObject(obj *object) error {
	switch obj.Type.Kind() {

	case reflect.Slice:
		if isInterfaceCollection(obj.Type) {
			dbg(2, "Slice", "interface collection")

			dep, err := v.storage.Get(obj.Address)
			if err == nil && dep.Value.IsValid() {
				dbg(3, "check", "already initialized")
				return nil
			}

			if err = v.storage.Set(&object{
				Address: obj.Address,
				Type:    obj.Type,
				Value:   v.storage.GetCollection(obj.Type),
			}); err != nil {
				dbg(3, "err", err)
				return err
			}
		} else {
			dbg(2, "Slice", "skip")
		}

		return nil

	case reflect.Func:
		dbg(2, "Func", "in", obj.Type.NumIn(), "out", obj.Type.NumOut())

		if !obj.Value.IsValid() {
			dbg(3, "validate", "got nil object value")
			return fmt.Errorf("got nil object value")
		}

		args := make([]reflect.Value, 0, obj.Type.NumIn())
		for i := 0; i < obj.Type.NumIn(); i++ {
			inRefType := obj.Type.In(i)
			inAddress := ResolveAddress(inRefType, reflect.Value{})

			dbg(3, "check in", inAddress)
			dep, err := v.storage.Get(inAddress)
			if err != nil {
				dbg(4, "err", err)
				return err
			}

			if !dep.Value.IsValid() {
				dbg(4, "validate", "not initialized")
				return fmt.Errorf("dependency [%s] not initialized", inAddress)
			}

			args = append(args, dep.Value)
		}

		dbg(3, "call", obj.Address)

		args = obj.Value.Call(args)

		for i, arg := range args {
			returnType := obj.Type.Out(i)
			outAddress := ResolveAddress(returnType, arg)
			dbg(3, "check out", outAddress)

			if isError(returnType) {
				if err, ok := arg.Interface().(error); ok && err != nil {
					dbg(4, "err", err)
					return err
				}
				dbg(4, "err", "nil")
				continue
			}

			dbg(3, "save", outAddress)

			err := v.storage.Set(&object{
				Address: outAddress,
				Type:    returnType,
				Value:   arg,
			})
			if err != nil {
				dbg(4, "err", err)
				return err
			}
		}

		return nil

	case reflect.Struct:
		dbg(2, "Struct", "fields", obj.Type.NumField())
		value := reflect.New(obj.Type)

		for i := 0; i < obj.Type.NumField(); i++ {
			inRefType := obj.Type.Field(i)
			inAddress := ResolveAddress(inRefType.Type, reflect.Value{})
			dbg(3, inRefType.Name, inAddress)

			if !inRefType.IsExported() {
				dbg(4, "skip", "private")
				continue
			}

			dep, err := v.storage.Get(inAddress)
			if err != nil {
				dbg(4, "err", err)
				return err
			}
			if !dep.Value.IsValid() {
				dbg(4, "check", "not initialized")
				return fmt.Errorf("dependency [%s] not initialized", inAddress)
			}

			value.Elem().FieldByName(inRefType.Name).Set(dep.Value)
		}

		obj.Value = value.Elem()

		return nil

	default:
		if !obj.Value.IsValid() {
			dbg(2, "validate", "got nil object value")
			return fmt.Errorf("got nil object value")
		}

		dbg(2, "any", "skip")
		return nil
	}
}
