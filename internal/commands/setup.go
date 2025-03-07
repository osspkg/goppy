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
			fmt.Fprintf(buff, "\n%s\n", datum)
		}
		return buff.Bytes(), nil
	}), "update .gitignore")
}

var tools1 = map[string]string{
	"goveralls":   "go install github.com/mattn/goveralls@latest",
	"static":      "go install go.osspkg.com/static/cmd/static@latest",
	"easyjson":    "go install github.com/mailru/easyjson/...@latest",
	"govulncheck": "go install golang.org/x/vuln/cmd/govulncheck@latest",
	"goimports":   "go install golang.org/x/tools/cmd/goimports@latest",
}

var tools2 = map[string]map[string]string{
	"go1.24": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5",
	},
	"go1.23": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5",
	},
	"go1.22": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2",
	},
	"go1.21": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0",
	},
	"go1.20": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.0",
	},
	"go1.19": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1",
	},
	"go1.18": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.3",
	},
	"go1.17": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2",
	},
}

var ciConfigs = map[string]string{
	".golangci.yml":            golangciLintConfig,
	"Makefile":                 makefileConfig,
	".github/workflows/ci.yml": githubCiConfig,
	".github/dependabot.yml":   githubDependabotConfig,
}

var golangciLintConfig = `
run:
  go: "{%go_ver%}"
  concurrency: 4
  timeout: 5m
  tests: false
  issues-exit-code: 1
  modules-download-mode: readonly

issues:
  exclude-use-default: false
  max-issues-per-linter: 100
  max-same-issues: 4
  new: false
  exclude-files:
    - ".+_test.go"
  exclude-dirs:
    - "vendor$"

output:
  formats:
    - format: line-number
  sort-results: true

linters-settings:
  govet:
    check-shadowing: true
    enable:
      - asmdecl
      - assign
      - atomic
      - atomicalign
      - bools
      - buildtag
      - cgocall
      - composites
      - copylocks
      - deepequalerrors
      - errorsas
      - findcall
      - framepointer
      - httpresponse
      - ifaceassert
      - loopclosure
      - lostcancel
      - nilfunc
      - nilness
      - printf
      - reflectvaluecompare
      - shadow
      - shift
      - sigchanyzer
      - sortslice
      - stdmethods
      - stringintconv
      - structtag
      - testinggoroutine
      - tests
      - unmarshal
      - unreachable
      - unsafeptr
      - unusedresult
      - unusedwrite
    disable:
      - fieldalignment
  gofmt:
    simplify: true
  errcheck:
    check-type-assertions: true
    check-blank: true
  gocyclo:
    min-complexity: 30
  misspell:
    locale: US
  prealloc:
    simple: true
    range-loops: true
    for-loops: true
  unparam:
    check-exported: false
  gci:
    skip-generated: true
    custom-order: false
  gosec:
    includes:
      - G101 # Look for hard coded credentials
      - G102 # Bind to all interfaces
      - G103 # Audit the use of unsafe block
      - G104 # Audit errors not checked
      - G106 # Audit the use of ssh.InsecureIgnoreHostKey
      - G107 # Url provided to HTTP request as taint input
      - G108 # Profiling endpoint automatically exposed on /debug/pprof
      - G109 # Potential Integer overflow made by strconv.Atoi result conversion to int16/32
      - G110 # Potential DoS vulnerability via decompression bomb
      - G111 # Potential directory traversal
      - G112 # Potential slowloris attack
      - G113 # Usage of Rat.SetString in math/big with an overflow (CVE-2022-23772)
      - G114 # Use of net/http serve function that has no support for setting timeouts
      - G201 # SQL query construction using format string
      - G202 # SQL query construction using string concatenation
      - G203 # Use of unescaped data in HTML templates
      - G204 # Audit use of command execution
      - G301 # Poor file permissions used when creating a directory
      - G302 # Poor file permissions used with chmod
      - G303 # Creating tempfile using a predictable path
      - G304 # File path provided as taint input
      - G305 # File traversal when extracting zip/tar archive
      - G306 # Poor file permissions used when writing to a new file
      - G307 # Deferring a method which returns an error
      - G401 # Detect the usage of DES, RC4, MD5 or SHA1
      - G402 # Look for bad TLS connection settings
      - G403 # Ensure minimum RSA key length of 2048 bits
      - G404 # Insecure random number source (rand)
      - G501 # Import blocklist: crypto/md5
      - G502 # Import blocklist: crypto/des
      - G503 # Import blocklist: crypto/rc4
      - G504 # Import blocklist: net/http/cgi
      - G505 # Import blocklist: crypto/sha1
      - G601 # Implicit memory aliasing of items from a range statement
    excludes:
      - G101 # Look for hard coded credentials
      - G102 # Bind to all interfaces
      - G103 # Audit the use of unsafe block
      - G104 # Audit errors not checked
      - G106 # Audit the use of ssh.InsecureIgnoreHostKey
      - G107 # Url provided to HTTP request as taint input
      - G108 # Profiling endpoint automatically exposed on /debug/pprof
      - G109 # Potential Integer overflow made by strconv.Atoi result conversion to int16/32
      - G110 # Potential DoS vulnerability via decompression bomb
      - G111 # Potential directory traversal
      - G112 # Potential slowloris attack
      - G113 # Usage of Rat.SetString in math/big with an overflow (CVE-2022-23772)
      - G114 # Use of net/http serve function that has no support for setting timeouts
      - G201 # SQL query construction using format string
      - G202 # SQL query construction using string concatenation
      - G203 # Use of unescaped data in HTML templates
      - G204 # Audit use of command execution
      - G301 # Poor file permissions used when creating a directory
      - G302 # Poor file permissions used with chmod
      - G303 # Creating tempfile using a predictable path
      - G304 # File path provided as taint input
      - G305 # File traversal when extracting zip/tar archive
      - G306 # Poor file permissions used when writing to a new file
      - G307 # Deferring a method which returns an error
      - G401 # Detect the usage of DES, RC4, MD5 or SHA1
      - G402 # Look for bad TLS connection settings
      - G403 # Ensure minimum RSA key length of 2048 bits
      - G404 # Insecure random number source (rand)
      - G501 # Import blocklist: crypto/md5
      - G502 # Import blocklist: crypto/des
      - G503 # Import blocklist: crypto/rc4
      - G504 # Import blocklist: net/http/cgi
      - G505 # Import blocklist: crypto/sha1
      - G601 # Implicit memory aliasing of items from a range statement
    exclude-generated: true
    severity: medium
    confidence: medium
    concurrency: 12
    config:
      global:
        nosec: true
        "#nosec": "#my-custom-nosec"
        show-ignored: true
        audit: true
      G101:
        pattern: "(?i)passwd|pass|password|pwd|secret|token|pw|apiKey|bearer|cred"
        ignore_entropy: false
        entropy_threshold: "80.0"
        per_char_threshold: "3.0"
        truncate: "32"
      G104:
        fmt:
          - Fscanf
      G111:
        pattern: "http\\.Dir\\(\"\\/\"\\)|http\\.Dir\\('\\/'\\)"
      G301: "0750"
      G302: "0600"
      G306: "0600"

  lll:
    line-length: 130
    tab-width: 1
  staticcheck:
    go: "1.15"
    # SAxxxx checks in https://staticcheck.io/docs/configuration/options/#checks
    # Default: ["*"]
    checks: [ "*", "-SA1019" ]

linters:
  disable-all: true
  enable:
    - govet
    - gofmt
    - errcheck
    - misspell
    - gocyclo
    - ineffassign
    - goimports
    - nakedret
    - unparam
    - unused
    - prealloc
    - durationcheck
    - staticcheck
    - makezero
    - nilerr
    - errorlint
    - bodyclose
    - exportloopref
    - gosec
    - lll
  fast: false
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
