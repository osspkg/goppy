/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"fmt"
	"log"
	"net/http"

	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/console"
)

func CmdFS() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("fs", "Run file server")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.IntVar("port", 8080, "set port")
		})
		setter.ExecFunc(func(port int64) {
			serv := http.FileServer(http.Dir(fs.CurrentDir()))
			http.Handle("/", serv)
			log.Println("FS running on http://localhost:8080")
			log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)) //nolint:gosec
		})
	})
}
