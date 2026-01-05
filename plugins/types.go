/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import "go.osspkg.com/xc"

type Broker interface {
	Name() string
	Priority() int
	Apply(arg any)
	OnStart(ctx xc.Context) error
	OnStop() error
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
