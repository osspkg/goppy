/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package main

import (
	"go.osspkg.com/goppy"
	"go.osspkg.com/goppy/ormmysql"
)

func main() {

	app := goppy.New()
	app.Plugins(
		ormmysql.WithMySQL(),
	)
	app.Run()

}
