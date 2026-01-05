/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package applog

type DevNull struct {
}

func (*DevNull) Write(p []byte) (int, error) {
	return len(p), nil
}
