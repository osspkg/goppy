/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

type (
	// Kind plugin structure
	Kind struct {
		Config  any
		Inject  any
		Resolve any
	}

	Kinds []Kind
)

func (p Kinds) Inject(list ...any) Kinds {
	for _, vv := range list {
		switch v := vv.(type) {
		case Kinds:
			p = append(p, v...)
		case Kind:
			p = append(p, v)
		default:
			p = append(p, Kind{Inject: vv})
		}
	}
	return p
}

func Inject(list ...any) Kinds {
	return (make(Kinds, 0, len(list))).Inject(list...)
}

// Defaulter interface for setting default values for a structure
type Defaulter interface {
	Default()
}

// Defaulter2 interface for setting default values for a structure
type Defaulter2 interface {
	Default() error
}

// Validator config validate
type Validator interface {
	Validate() error
}
