/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v2/internal/global"
)

func CmdSetupLib() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("setup-lib", "Setup lib project")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.Bool("force", "force update")
		})
		setter.ExecFunc(func(_ []string, force bool) {
			console.Infof("--- SETUP LIB ---")

			updateGitIgnore()
			installTools()
			addCICD(force)
			updateGoMod()
		})
	})
}

func CmdSetupApp() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("setup-app", "Setup app project")
		setter.Flag(func(flagsSetter console.FlagsSetter) {
			flagsSetter.Bool("force", "force update")
		})
		setter.ExecFunc(func(_ []string, force bool) {
			console.Infof("--- SETUP APP ---")

			updateGitIgnore()
			installTools()
			addCICD(force)
			updateGoMod()

			createAppDirs()
			createScripts(force)

		})
	})
}

func createScripts(force bool) {
	console.Infof("create services and deb scripts")
	postinstData, postrmData, preinstData, prermData := bashPrefix, bashPrefix, bashPrefix, bashPrefix

	mainFiles, err := fs.SearchFiles(fs.CurrentDir(), "main.go")
	console.FatalIfErr(err, "detect main.go")
	for _, main := range mainFiles {
		appName := fs.DirName(main)
		repl := strings.NewReplacer(
			"{%app_name%}", appName,
		)
		if !fs.FileExist(global.GetInitDir()+"/"+appName+".service") || force {
			tmpl := repl.Replace(systemctlConfig)
			console.FatalIfErr(
				os.WriteFile(global.GetInitDir()+"/"+appName+".service", []byte(tmpl), 0755),
				"create init config [%s]", appName)
		}

		postinstData += repl.Replace(postinst)
		preinstData += repl.Replace(preinstDir)
		preinstData += repl.Replace(preinst)
		prermData += repl.Replace(prerm)
	}

	files := map[string]string{
		"postinst.sh": postinstData,
		"postrm.sh":   postrmData,
		"preinst.sh":  preinstData,
		"prerm.sh":    prermData,
	}
	scriptsDir := global.GetScriptsDir()
	for fileName, fileValue := range files {
		filePath := scriptsDir + "/" + fileName
		if !fs.FileExist(filePath) || force {
			console.FatalIfErr(os.WriteFile(filePath, []byte(fileValue), 0755), "create postinst")
		}
	}
}

func createAppDirs() {
	console.FatalIfErr(os.MkdirAll(global.GetInitDir(), 0755), "create init dir")
	console.FatalIfErr(os.MkdirAll(global.GetScriptsDir(), 0755), "create scripts dir")
}

func addCICD(force bool) {
	repl := strings.NewReplacer(
		"{%go_ver%}", strings.Trim(global.GoVersion(), "go"),
	)
	console.Infof("create ci/cd configs")
	for name, config := range ciConfigs {
		if !force && fs.FileExist(fs.CurrentDir()+"/"+name) {
			continue
		}
		if strings.Contains(name, "/") {
			console.FatalIfErr(os.MkdirAll(fs.CurrentDir()+"/"+filepath.Dir(name), 0755), "create dir for [%s]", name)
		}
		config = repl.Replace(config)
		console.FatalIfErr(os.WriteFile(fs.CurrentDir()+"/"+name, []byte(config), 0744), "create config [%s]", name)
	}
}

func installTools() {
	toolDir := global.GetToolsDir()
	console.FatalIfErr(os.MkdirAll(toolDir, 0755), "create tools dir")

	console.Infof("install tools")
	for name, install := range tools1 {
		if !fs.FileExist(toolDir + "/" + name) {
			global.ExecPack(install)
		}
	}

	goVersion := global.GoVersion()
	console.Infof("go version: %s", goVersion)
	tools, ok := tools2[goVersion]
	if ok {
		for name, install := range tools {
			if !fs.FileExist(toolDir + "/" + name) {
				global.ExecPack(install)
			}
		}
	}
}

func updateGoMod() {
	cmds := make([]string, 0, 50)
	cmds = append(cmds, "go version")

	if fs.FileExist(fs.CurrentDir() + "/go.work") {
		cmds = append(cmds, "go work use -r .", "go work sync")
		mods, err := fs.SearchFiles(fs.CurrentDir(), "go.mod")
		console.FatalIfErr(err, "detects go.mod in workspace")
		for _, mod := range mods {
			dir := filepath.Dir(mod)
			cmds = append(cmds,
				"cd "+dir+" && go mod tidy -v -compat=1.17 && go mod download",
			)
		}
	} else {
		cmds = append(cmds,
			"go mod tidy -v -compat=1.17",
			"go mod download",
		)
	}

	global.ExecPack(cmds...)
}

func updateGitIgnore() {
	console.Infof("update .gitignore")
	console.FatalIfErr(fs.RewriteFile(fs.CurrentDir()+"/.gitignore", func(b []byte) ([]byte, error) {
		buff := bytes.NewBuffer(b)
		data := []string{
			global.DirTools + "/", "bin/", "vendor/", "build/",
			".idea/", ".vscode/",
			"coverage.txt", "coverage.out",
			"*.exe", "*.exe~", "*.dll", "*.so", "*.dylib", "*.db", "*.db-journal",
			"*.mmdb", "*.test", "*.out", ".env",
		}
		for _, datum := range data {
			if bytes.Contains(b, []byte(datum)) {
				continue
			}
			fmt.Fprintf(buff, "\n%s", datum)
		}
		return buff.Bytes(), nil
	}), "update .gitignore")
}

var tools1 = map[string]string{
	"goveralls":     "go install github.com/mattn/goveralls@latest",
	"static":        "go install go.osspkg.com/static/cmd/static@latest",
	"easyjson":      "go install github.com/mailru/easyjson/...@latest",
	"govulncheck":   "go install golang.org/x/vuln/cmd/govulncheck@latest",
	"goimports":     "go install golang.org/x/tools/cmd/goimports@latest",
	"golangci-lint": "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.2",
}

var tools2 = map[string]map[string]string{
	"go1.24": {
		//"golangci-lint": "go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.2",
	},
}

var ciConfigs = map[string]string{
	".golangci.yml":            golangciLintConfig,
	"Makefile":                 makefileConfig,
	".github/workflows/ci.yml": githubCiConfig,
	".github/dependabot.yml":   githubDependabotConfig,
}

var golangciLintConfig = `version: "2"

run:
  go: "{%go_ver%}"
  timeout: 5m
  tests: false
  issues-exit-code: 1
  modules-download-mode: readonly
  allow-parallel-runners: true

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
  new: false
  fix: false

output:
  formats:
    text:
      print-linter-name: true
      print-issued-lines: true

formatters:
  exclusions:
    paths:
      - vendors/
  enable:
    - gofmt
    - goimports

linters:
  settings:
    staticcheck:
      checks:
        - all
        - -S1023
        - -ST1000
        - -ST1003
        - -ST1020
    gosec:
      excludes:
        - G104
        - G115
        - G301
        - G304
        - G306
        - G501
        - G505
  exclusions:
    paths:
      - vendors/
  default: none
  enable:
    - govet
    - errcheck
    - misspell
    - gocyclo
    - ineffassign
    - unparam
    - unused
    - prealloc
    - durationcheck
    - staticcheck
    - makezero
    - nilerr
    - errorlint
    - bodyclose
    - gosec
`

var makefileConfig = `
SHELL=/bin/bash


.PHONY: install
install:
	go install go.osspkg.com/goppy/v2/cmd/goppy@latest
	goppy setup-lib

.PHONY: lint
lint:
	goppy lint

.PHONY: license
license:
	goppy license

.PHONY: build
build:
	goppy build --arch=amd64

.PHONY: tests
tests:
	goppy test

.PHONY: pre-commit
pre-commit: install license lint tests build

.PHONY: ci
ci: pre-commit

`

var systemctlConfig = `[Unit]
After=network.target

[Service]
User=root
Group=root
Restart=on-failure
RestartSec=30s
Type=simple
ExecStart=/usr/bin/{%app_name%} --config=/etc/{%app_name%}/config.yaml
KillMode=process
KillSignal=SIGTERM

[Install]
WantedBy=default.target
`

var (
	bashPrefix = "#!/bin/bash\n\n"
	postinst   = `
if [ -f "/etc/systemd/system/{%app_name%}.service" ]; then
    systemctl start {%app_name%}
    systemctl enable {%app_name%}
    systemctl daemon-reload
fi
`
	preinstDir = `
if ! [ -d /var/lib/{%app_name%}/ ]; then
    mkdir /var/lib/{%app_name%}
fi
`
	preinst = `
if [ -f "/etc/systemd/system/{%app_name%}.service" ]; then
    systemctl stop {%app_name%}
    systemctl disable {%app_name%}
    systemctl daemon-reload
fi
`
	prerm = `
if [ -f "/etc/systemd/system/{%app_name%}.service" ]; then
    systemctl stop {%app_name%}
    systemctl disable {%app_name%}
    systemctl daemon-reload
fi
`
)

var githubCiConfig = `
name: CI

on:
  push:
    branches: [ master ]
  pull_request:
    branches: [ master ]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go: [ '{%go_ver%}' ]
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: ${{ matrix.go }}

      - name: Run CI
        env:
          COVERALLS_TOKEN: ${{ secrets.COVERALLS_TOKEN }}
        run: make ci
`

var githubDependabotConfig = `
version: 2
updates:
  - package-ecosystem: "gomod" # See documentation for possible values
    directory: "/" # Location of package manifests
    schedule:
      interval: "weekly"
`
