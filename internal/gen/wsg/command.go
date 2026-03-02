/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package wsg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go.osspkg.com/do"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/builder"
	"go.osspkg.com/goppy/v3/internal/global"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/syncing"
	"golang.org/x/mod/modfile"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/internal/gen/wsg/visitor"
)

func Command() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("wsg", "generate web server api")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.String("out", "output file specified")
			flagsSetter.StringVar("iface", "", "interface names (optional)")
		})
		setter.ExecFunc(func(out, _iface string) {
			console.ShowDebug(true)
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

			iface := do.Entries[string, string, struct{}](strings.Split(_iface, ","), func(s string) (string, struct{}) {
				return strings.ToLower(strings.TrimSpace(s)), struct{}{}
			})

			build := &builder.Builder{
				Out:   out,
				IFace: iface,
			}

			for _, filePath := range files {
				if global.NeedSkipFile(filePath) {
					continue
				}

				console.Debugf("> PARSE FILE: %s", filePath)

				gomod, root, err := detectGoMod(filePath)
				console.FatalIfErr(err, "failed to detect gomod")

				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
				console.FatalIfErr(err, "parse go file: %s", filePath)

				if ast.IsGenerated(f) {
					continue
				}

				vv := &visitor.Visitor{
					Imports:  syncing.NewMap[string, string](1),
					FilePath: strings.TrimPrefix(filePath, root),
					GoMod:    gomod,
				}

				ast.Walk(vv, f)

				vv.Debug()

				build.Files = append(build.Files, vv.ToFile())
			}

			console.FatalIfErr(build.Build(), "")
		})
	})
}

func detectGoMod(curPath string) (mod string, root string, err error) {
	root = filepath.Dir(curPath)
	for {

		mods, e := fs.SearchFilesByExt(root, ".mod")
		if e != nil {
			return "", "", e
		}

		if len(mods) != 0 {
			for _, s := range mods {
				b, e := os.ReadFile(s)
				if e != nil {
					return "", "", e
				}

				f, e := modfile.Parse("go.mod", b, nil)
				if e != nil {
					return "", "", e
				}
				mod = f.Module.Mod.Path
				return
			}
		}

		root, err = filepath.Abs(root + "/..")
		if err != nil {
			return
		}

		if root == "/" {
			err = fmt.Errorf("failed to detect go.mod")
			return
		}
	}
}
