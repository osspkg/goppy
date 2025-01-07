/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package env

import "os"

func Get(key, def string) string {
	if v, ok := os.LookupEnv(key); !ok {
		return v
	}
	return def
}
