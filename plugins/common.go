/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import (
	"fmt"
	"os"
	"reflect"
)

type (
	// Plugin plugin structure
	Plugin struct {
		Config  interface{}
		Inject  interface{}
		Resolve interface{}
	}

	Plugins []Plugin
)

func (p Plugins) Inject(list ...interface{}) Plugins {
	for _, vv := range list {
		switch v := vv.(type) {
		case Plugins:
			p = append(p, v...)
		case Plugin:
			p = append(p, v)
		default:
			switch reflect.TypeOf(vv).Kind() {
			case reflect.Ptr, reflect.Func:
				p = append(p, Plugin{Inject: vv})
			default:
				fmt.Printf("Plugins Inject error: unknown dependency %T", vv)
				os.Exit(1)
			}
		}
	}
	return p
}

// Defaulter interface for setting default values for a structure
type Defaulter interface {
	Default()
}

type Validator interface {
	Validate() error
}
