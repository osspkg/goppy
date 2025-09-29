/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlite

import "go.osspkg.com/goppy/v2/orm/dialect"

const Name dialect.Name = "sqlite"

var (
	_ dialect.Connector       = (*pool)(nil)
	_ dialect.ConfigInterface = (*ConfigGroup)(nil)
)

func init() {
	dialect.Register(Name, New(), migrate{})
}
