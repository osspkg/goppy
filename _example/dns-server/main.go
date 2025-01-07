/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy/v2"
	"go.osspkg.com/goppy/v2/xdns"
)

func main() {
	app := goppy.New("", "", "")
	app.Plugins(
		xdns.WithServer(),
	)
	app.Run()
}
