/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

type Option func(o *Migrate)

func UseMigration(m []Migration) Option {
	return func(o *Migrate) {
		o.FS = newMemFS(m)
	}
}