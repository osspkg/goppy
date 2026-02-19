/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package wsg

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/visitor"
)

func Command() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("wsg", "generate web server api")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.String("out", "output file specified")
		})
		setter.ExecFunc(func(out string) {
			//console.ShowDebug(false)
			console.Infof("--- GENERATE ---")

			curDir := fs.CurrentDir()

			console.FatalIfTrue(len(out) == 0, "no output file specified")
			console.FatalIfTrue(out == curDir, "output dir equals current dir")

			console.FatalIfErr(os.RemoveAll(out), "failed to remove old output directory")
			console.FatalIfErr(os.MkdirAll(out, 0755), "failed to create directory")

			files, err := fs.SearchFilesByExt(curDir, ".go")
			console.FatalIfErr(err, "search files in %s", curDir)

			files = do.Filter[string](files, func(value string, index int) bool {
				return !strings.HasPrefix(value, out)
			})

			for _, filePath := range files {
				console.Debugf("> PARSE FILE: %s", filePath)

				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
				console.FatalIfErr(err, "parse go file: %s", filePath)

				if ast.IsGenerated(f) {
					continue
				}

				vv := &visitor.Visitor{
					Imports:  syncing.NewMap[string, string](1),
					FilePath: filePath,
				}

				ast.Walk(vv, f)

				vv.Debug()
			}
		})
	})
}
