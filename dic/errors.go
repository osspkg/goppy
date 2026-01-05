/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"go.osspkg.com/errors"
)

var (
	ErrDepAlreadyRunning = errors.New("dependencies are already running")
	ErrDepNotRunning     = errors.New("dependencies are not running yet")
	ErrBreakPointType    = errors.New("breakpoint can only be a function")
	ErrBrokerExist       = errors.New("broker already exist")
	ErrInvokeType        = errors.New("invoke supported func only")
)
