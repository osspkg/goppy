/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

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
			p = append(p, Plugin{Inject: vv})
		}
	}
	return p
}

func Inject(list ...interface{}) Plugins {
	return Plugins{}.Inject(list...)
}

// Defaulter interface for setting default values for a structure
type Defaulter interface {
	Default()
}

type Validator interface {
	Validate() error
}
