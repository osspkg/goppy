/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"os"
	"strings"

	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/goppy/v3/apigen/builder"
	"go.osspkg.com/goppy/v3/apigen/parser"
	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/internal/global"
	"go.osspkg.com/ioutils/fs"
)

func CmdWSG() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("wsg", "generate web server api")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.String("out", "output file specified")
			flagsSetter.StringVar("iface", "", "interface names (optional)")
			flagsSetter.StringVar("mod", "json-rpc", "generation modules (optional)")
			flagsSetter.StringVar("pool", "main", "web server pool list (optional)")
		})
		setter.ExecFunc(func(out, _iface, _mod, _pool string) {
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

			mods := do.Treat[string](strings.Split(_mod, ","), func(value string, index int) string {
				return strings.ToLower(strings.TrimSpace(value))
			})

			pool := do.Treat[string](strings.Split(_pool, ","), func(value string, index int) string {
				return strings.ToLower(strings.TrimSpace(value))
			})

			face := do.Entries[string, string, struct{}](
				do.Filter[string](strings.Split(_iface, ","),
					func(value string, index int) bool {
						return len(value) > 0
					},
				),
				func(s string) (string, struct{}) {
					return strings.ToLower(strings.TrimSpace(s)), struct{}{}
				},
			)

			build := &builder.Builder{
				Out:   out,
				IFace: face,
				Mods:  mods,
				Pool:  pool,
			}

			for _, filePath := range files {
				if global.NeedSkipFile(filePath) {
					continue
				}

				console.Debugf("> PARSE FILE: %s", filePath)

				vv, e := parser.New("@wsg", filePath)
				if e != nil {
					if errors.Is(e, parser.ErrIsGenerated) {
						continue
					}
					console.FatalIfErr(e, "parse go file: %s", filePath)
				}

				vv.DumpStdout()
				build.Files = append(build.Files, vv.ToFile())
			}

			console.FatalIfErr(build.Build(), "")
		})
	})
}
