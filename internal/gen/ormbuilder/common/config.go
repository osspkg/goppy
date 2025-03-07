/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package common

import "go.osspkg.com/goppy/v2/internal/gen/ormbuilder/dialects"

type Config struct {
	Dialect         dialects.Dialect
	DBRead, DBWrite string
	Dir, SQLDir     string
	FileIndex       int64
}
