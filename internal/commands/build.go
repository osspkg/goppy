/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/goppy/v2/internal/global"
	"go.osspkg.com/ioutils/fs"
)

func CmdBuild() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("build", "Building app")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.StringVar("arch", "amd64,arm64", "")
			flagsSetter.StringVar("mode", "app", "")
		})
		setter.ExecFunc(func(_ []string, arch, mode string) {
			console.Infof("--- BUILD ---")

			pack := make([]string, 0, 10)
			buildDir := global.GetBuildDir()

			mainFiles, err := fs.SearchFiles(fs.CurrentDir(), "main.go")
			console.FatalIfErr(err, "detect main.go")

			for _, main := range mainFiles {
				appName := fs.ParentFolder(main)
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

			global.ExecPack(pack...)
		})
	})
}
