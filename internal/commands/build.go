/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/codec"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/console"
	"go.osspkg.com/goppy/v3/internal/global"
)

const buildConfigFileName = ".build.yaml"

type buildConfig struct {
	Mode string `yaml:"mode"`
	Arch string `yaml:"arch"`
	CGO  bool   `yaml:"cgo"`
}

func CmdBuild() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("build", "Building app")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.StringVar("arch", "amd64", "set architecture (amd64,arm64,js)")
			flagsSetter.StringVar("mode", "app", "set application mode (app,plugin,wasm)")
			flagsSetter.StringVar("main", "", "set main package for build")
			flagsSetter.Bool("cgo", "set CGO_ENABLED=1 for build")
		})
		setter.ExecFunc(func(_ []string, _arch, _mode, _main string, _cgo bool) {
			console.Infof("goppy build")

			pack := make([]string, 0, 10)
			buildDir := global.GetBuildDir()
			console.FatalIfErr(os.MkdirAll(buildDir, 0744), "create build dir")

			var mainFiles []string
			var err error

			if mb := global.SplitFirstString(",", _main); len(mb) > 0 {
				mainFiles = do.Filter(mb, func(value string, _ int) bool {
					return strings.HasSuffix(value, "/main.go")
				})
			} else {
				mainFiles, err = fs.SearchFiles(fs.CurrentDir(), "main.go")
				console.FatalIfErr(err, "detect main.go")
			}

			for _, main := range mainFiles {
				appName := fs.DirName(main)
				switch appName {
				case "", "/", ".":
					appName = "app"
				}

				console.Infof("build[%s]: %s", appName, main)

				var config buildConfig
				configPath := filepath.Join(filepath.Dir(main), buildConfigFileName)
				if fs.FileExist(configPath) {
					console.Warnf("build config already exists: %s", configPath)
					console.FatalIfErr(codec.FileEncoder(configPath).Decode(&config), "build config")
				}

				modeList := global.SplitFirstString(",", config.Mode, _mode)
				archList := global.SplitFirstString(",", config.Arch, _arch)
				cgo := _cgo || config.CGO

				for _, mode := range modeList {
					for _, arch := range archList {

						chunk := []string{
							"GODEBUG=netdns=9",
							"GO111MODULE=on",
						}

						switch arch {
						case "arm64":
							chunk = append(chunk, "GOOS=linux", "GOARCH=arm64")
							if cgo {
								if fs.FileExist("/usr/bin/aarch64-linux-gnu-gcc") {
									chunk = append(chunk, "CC=aarch64-linux-gnu-gcc")
								} else {
									console.Warnf("please install aarch64-linux-gnu-gcc")
								}
							}

						case "amd64":
							chunk = append(chunk, "GOOS=linux", "GOARCH=amd64")

						case "js":
							chunk = append(chunk, "GOOS=js", "GOARCH=wasm")
							mode = "wasm"

						default:
							console.Warnf("possible only arch: amd64,arm64,js")
							continue
						}

						switch mode {
						case "app":
							chunk = append(chunk,
								"CGO_ENABLED="+do.IfElse(cgo, "1", "0"),
								`go build -ldflags='-s -w' -a -o `+buildDir+"/"+appName+"_"+arch+" "+main)

						case "plugin":
							chunk = append(chunk,
								"CGO_ENABLED="+do.IfElse(cgo, "1", "0"),
								`go build -buildmode=plugin -ldflags='-s -w' -a -o `+buildDir+"/"+appName+"_"+arch+".so "+main)

						case "wasm":
							chunk = append(chunk,
								"CGO_ENABLED=0",
								`go build -ldflags='-s -w' -a -o `+buildDir+"/"+appName+".wasm "+main)
							pack = append(pack, "cp -rf \"$(go env GOROOT)/lib/wasm/wasm_exec.js\" "+buildDir+"/wasm_exec.js")
							console.FatalIfErr(os.WriteFile(
								buildDir+"/"+appName+".html",
								[]byte(fmt.Sprintf(htmlWasmTmpl, appName)),
								0644), "write html")

						default:
							console.Warnf("possible only mode: app,plugin,wasm")
							continue
						}

						pack = append(pack, "rm -rf "+buildDir+"/"+appName+"_"+arch)
						pack = append(pack, strings.Join(chunk, " "))
					}
				}
			}

			global.ExecPack(true, pack...)
		})
	})
}

const htmlWasmTmpl = `<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="utf-8">
        <script src="wasm_exec.js"></script>
        <script>
            const go = new Go();
            WebAssembly.instantiateStreaming(fetch("%s.wasm"), go.importObject).then((result) => {
                go.run(result.instance);
            });
        </script>
        <title>Wasm App</title>
    </head>
    <body></body>
</html>`
