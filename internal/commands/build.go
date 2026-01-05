/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/internal/global"
)

func CmdBuild() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("build", "Building app")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.StringVar("arch", "amd64,arm64", "set architecture")
			flagsSetter.StringVar("mode", "app", "set application mode (app, plugin)")
			flagsSetter.StringVar("main", "", "set main package for build")
		})
		setter.ExecFunc(func(_ []string, arch, mode, _main string) {
			console.Infof("--- BUILD ---")

			pack := make([]string, 0, 10)
			buildDir := global.GetBuildDir()

			mainFiles, err := fs.SearchFiles(fs.CurrentDir(), "main.go")
			console.FatalIfErr(err, "detect main.go")

			mb := do.TreatValue(strings.Split(_main, ","), strings.ToLower, strings.TrimSpace)
			if len(mb) > 0 {
				mainFiles = do.Filter(mainFiles, func(value string, index int) bool {
					for _, s := range mb {
						if strings.HasSuffix(value, s+"/main.go") {
							return true
						}
					}
					return false
				})
			}

			for _, main := range mainFiles {
				appName := fs.DirName(main)
				archList := strings.Split(arch, ",")

				for _, arch = range archList {
					pack = append(pack, "rm -rf "+buildDir+"/"+appName+"_"+arch)

					chunk := []string{
						"GODEBUG=netdns=9",
						"GO111MODULE=on",
						"CGO_ENABLED=1",
					}

					switch arch {
					case "arm64":
						chunk = append(chunk, "GOOS=linux", "GOARCH=arm64")
						if fs.FileExist("/usr/bin/aarch64-linux-gnu-gcc") {
							chunk = append(chunk, "CC=aarch64-linux-gnu-gcc")
						}
					case "amd64":
						chunk = append(chunk, "GOOS=linux", "GOARCH=amd64")
					default:
						console.Fatalf("possible only arch: amd64, arm64")
					}

					switch mode {
					case "app":
						chunk = append(chunk, `go build -ldflags='-s -w' -a -o `+buildDir+"/"+appName+"_"+arch+" "+main)
					case "plugin":
						chunk = append(chunk, `go build -buildmode=plugin -ldflags='-s -w' -a -o `+buildDir+"/"+appName+"_"+arch+".so "+main)
					default:
						console.Fatalf("possible only mode: app, plugin")
					}

					pack = append(pack, strings.Join(chunk, " "))
				}
			}

			global.ExecPack(true, pack...)
		})
	})
}
