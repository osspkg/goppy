/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"os"

	"go.osspkg.com/console"

	"go.osspkg.com/goppy/v2/internal/global"
)

func CmdTest() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("test", "Testing code")
		setter.ExecFunc(func(_ []string) {
			console.Infof("--- TESTS ---")

			pack := []string{
				"go clean -testcache",
				"go test -v -race -run Unit -covermode=atomic -coverprofile=coverage.out ./...",
			}

			coverallsToken := os.Getenv("COVERALLS_TOKEN")
			if len(coverallsToken) > 0 {
				pack = append(pack, "goveralls -coverprofile=coverage.out -repotoken "+coverallsToken)
			}

			global.ExecPack(true, pack...)
		})
	})
}
