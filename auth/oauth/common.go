/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package oauth

import "go.osspkg.com/errors"

var (
	ErrProviderNotFound = errors.New("provider not found")
)
