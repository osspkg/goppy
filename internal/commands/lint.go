/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"fmt"
	"path/filepath"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v2/internal/global"
)

func CmdLint() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("lint", "Linting code")
		setter.ExecFunc(func(_ []string) {
			console.Infof("--- LINT ---")

			updateGoMod()

			cmds := make([]string, 0, 50)
			cmds = append(cmds, "go generate ./...", "gofmt -w -s .", "golangci-lint --version")
			mods, err := fs.SearchFiles(fs.CurrentDir(), "go.mod")
			console.FatalIfErr(err, "detects go.mod in workspace")
			for _, mod := range mods {
				folder := filepath.Dir(mod)
				modName := global.GoModule(folder)
				cmds = append(cmds, fmt.Sprintf(
					"cd %s && go generate ./... && goimports -l -local=%s -w . && golangci-lint -v run ./...",
					folder, modName,
				))
			}

			global.ExecPack(true, cmds...)

			cmds = make([]string, 0, 50)
			cmds = append(cmds, "govulncheck ./...")

			global.ExecPack(false, cmds...)
		})
	})
}
