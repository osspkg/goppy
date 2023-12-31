/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package app

import "go.osspkg.com/goppy/errors"

var (
	errDepAlreadyRunned = errors.New("dependencies are already running")
	errDepNotRunning    = errors.New("dependencies are not running yet")
	errServiceUnknown   = errors.New("unknown service")
)
