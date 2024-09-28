/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"path/filepath"

	"go.osspkg.com/console"
	"go.osspkg.com/goppy/v2/internal/global"
	"go.osspkg.com/ioutils/fs"
)

func CmdLint() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("lint", "Linting code")
		setter.ExecFunc(func(_ []string) {
			console.Infof("--- LINT ---")

			updateGoMod()

			cmds := make([]string, 0, 50)
			cmds = append(cmds, "gofmt -w -s .", "golangci-lint --version")
			mods, err := fs.SearchFiles(fs.CurrentDir(), "go.mod")
			console.FatalIfErr(err, "detects go.mod in workspace")
			for _, mod := range mods {
				cmds = append(cmds, "cd "+filepath.Dir(mod)+" && golangci-lint -v run ./...")
			}

			global.ExecPack(cmds...)
		})
	})
}
