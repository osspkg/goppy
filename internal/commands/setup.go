/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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
	"go.osspkg.com/goppy/v2/internal/global"
	"go.osspkg.com/ioutils/fs"
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
		appName := fs.ParentFolder(main)
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
				"cd "+dir+" && go mod tidy -v -compat=1.17 && go mod download && go generate ./...",
			)
		}
	} else {
		cmds = append(cmds,
			"go mod tidy -v -compat=1.17",
			"go mod download",
			"go generate ./...",
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
			fmt.Fprintf(buff, "%s\n", datum)
		}
		return buff.Bytes(), nil
	}), "update .gitignore")
}

var tools1 = map[string]string{
	"goveralls": "go install github.com/mattn/goveralls@latest",
	"static":    "go install go.osspkg.com/static/cmd/static@latest",
	"easyjson":  "go install github.com/mailru/easyjson/...@latest",
}

var tools2 = map[string]map[string]string{
	"go1.23": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.2",
	},
	"go1.22": {
		"golangci-lint": "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.2",
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
# options for analysis running
run:
  # timeout for analysis, e.g. 30s, 5m, default is 1m
  deadline: 5m

  # exit code when at least one issue was found, default is 1
  issues-exit-code: 1

  # include test files or not, default is true
  tests: true

  # which files to skip: they will be analyzed, but issues from them
  # won't be reported. Default value is empty list, but there is
  # no need to include all autogenerated files, we confidently recognize
  # autogenerated files. If it's not please let us know.
  skip-files:
    - easyjson

issues:
  # Independently from option 'exclude' we use default exclude patterns,
  # it can be disabled by this option. To list all
  # excluded by default patterns execute 'golangci-lint run --help'.
  # Default value for this option is true.
  exclude-use-default: false
  # Excluding configuration per-path, per-linter, per-text and per-source
  exclude-rules:
    # Exclude some linters from running on tests files.
    - path: _test\.go
      linters:
        - prealloc
        - errcheck

# output configuration options
output:
  # colored-line-number|line-number|json|tab|checkstyle, default is "colored-line-number"
  format: colored-line-number

  # print lines of code with issue, default is true
  print-issued-lines: true

  # print linter name in the end of issue text, default is true
  print-linter-name: true

# all available settings of specific linters
linters-settings:
  govet:
    # report about shadowed variables
    check-shadowing: true
  enable:
    # report mismatches between assembly files and Go declarations
    - asmdecl
    # check for useless assignments
    - assign
    # check for common mistakes using the sync/atomic package
    - atomic
    # check for non-64-bits-aligned arguments to sync/atomic functions
    - atomicalign
    # check for common mistakes involving boolean operators
    - bools
    # check that +build tags are well-formed and correctly located
    - buildtag
    # detect some violations of the cgo pointer passing rules
    - cgocall
    # check for unkeyed composite literals
    - composites
    # check for locks erroneously passed by value
    - copylocks
    # check for calls of reflect.DeepEqual on error values
    - deepequalerrors
    # report passing non-pointer or non-error values to errors.As
    - errorsas
    # find calls to a particular function
    - findcall
    # report assembly that clobbers the frame pointer before saving it
    - framepointer
    # check for mistakes using HTTP responses
    - httpresponse
    # detect impossible interface-to-interface type assertions
    - ifaceassert
    # check references to loop variables from within nested functions
    - loopclosure
    # check cancel func returned by context.WithCancel is called
    - lostcancel
    # check for useless comparisons between functions and nil
    - nilfunc
    # check for redundant or impossible nil comparisons
    - nilness
    # check consistency of Printf format strings and arguments
    - printf
    # check for comparing reflect.Value values with == or reflect.DeepEqual
    - reflectvaluecompare
    # check for possible unintended shadowing of variables
    - shadow
    # check for shifts that equal or exceed the width of the integer
    - shift
    # check for unbuffered channel of os.Signal
    - sigchanyzer
    # check the argument type of sort.Slice
    - sortslice
    # check signature of methods of well-known interfaces
    - stdmethods
    # check for string(int) conversions
    - stringintconv
    # check that struct field tags conform to reflect.StructTag.Get
    - structtag
    # report calls to (*testing.T).Fatal from goroutines started by a test.
    - testinggoroutine
    # check for common mistaken usages of tests and examples
    - tests
    # report passing non-pointer or non-interface values to unmarshal
    - unmarshal
    # check for unreachable code
    - unreachable
    # check for invalid conversions of uintptr to unsafe.Pointer
    - unsafeptr
    # check for unused results of calls to some functions
    - unusedresult
    # checks for unused writes
    - unusedwrite
  disable:
    # find structs that would use less memory if their fields were sorted
    - fieldalignment
  gofmt:
    # simplify code: gofmt with '-s' option, true by default
    simplify: true
  errcheck:
    # report about not checking of errors in type assetions: 'a := b.(MyStruct)';
    # default is false: such cases aren't reported by default.
    check-type-assertions: true
    # report about assignment of errors to blank identifier: 'num, _ := strconv.Atoi(numStr)';
    # default is false: such cases aren't reported by default.
    check-blank: true
  gocyclo:
    # minimal code complexity to report, 30 by default (but we recommend 10-20)
    min-complexity: 15
  misspell:
    # Correct spellings using locale preferences for US or UK.
    # Default is to use a neutral variety of English.
    # Setting locale to US will correct the British spelling of 'colour' to 'color'.
    locale: US
  prealloc:
    # XXX: we don't recommend using this linter before doing performance profiling.
    # For most programs usage of prealloc will be a premature optimization.
    # Report preallocation suggestions only on simple loops that have no returns/breaks/continues/gotos in them.
    # True by default.
    simple: true
    range-loops: true # Report preallocation suggestions on range loops, true by default
    for-loops: true # Report preallocation suggestions on for loops, false by default
  unparam:
    # Inspect exported functions, default is false. Set to true if no external program/library imports your code.
    # XXX: if you enable this setting, unparam will report a lot of false-positives in text editors:
    # if it's called for subdir of a project it can't find external interfaces. All text editor integrations
    # with golangci-lint call it on a directory with the changed file.
    check-exported: false
  gci:
    # Section configuration to compare against.
    # Section names are case-insensitive and may contain parameters in ().
    # The default order of sections is 'standard > default > custom > blank > dot',
    # If 'custom-order' is 'true', it follows the order of 'sections' option.
    # Default: ["standard", "default"]
    #sections:
      #- standard # Standard section: captures all standard packages.
      #- default # Default section: contains all imports that could not be matched to another section type.
      #- blank # Blank section: contains all blank imports. This section is not present unless explicitly enabled.
      #- dot # Dot section: contains all dot imports. This section is not present unless explicitly enabled.
    # Skip generated files.
    # Default: true
    skip-generated: true
    # Enable custom order of sections.
    # If 'true', make the section order the same as the order of 'sections'.
    # Default: false
    custom-order: false
  gosec:
    # To select a subset of rules to run.
    # Available rules: https://github.com/securego/gosec#available-rules
    # Default: [] - means include all rules
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
    # To specify a set of rules to explicitly exclude.
    # Available rules: https://github.com/securego/gosec#available-rules
    # Default: []
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
    # Exclude generated files
    # Default: false
    exclude-generated: true
    # Filter out the issues with a lower severity than the given value.
    # Valid options are: low, medium, high.
    # Default: low
    severity: medium
    # Filter out the issues with a lower confidence than the given value.
    # Valid options are: low, medium, high.
    # Default: low
    confidence: medium
    # Concurrency value.
    # Default: the number of logical CPUs usable by the current process.
    concurrency: 12
    # To specify the configuration of rules.
    config:
      # Globals are applicable to all rules.
      global:
        # If true, ignore #nosec in comments (and an alternative as well).
        # Default: false
        nosec: true
        # Add an alternative comment prefix to #nosec (both will work at the same time).
        # Default: ""
        "#nosec": "#my-custom-nosec"
        # Define whether nosec issues are counted as finding or not.
        # Default: false
        show-ignored: true
        # Audit mode enables addition checks that for normal code analysis might be too nosy.
        # Default: false
        audit: true
      G101:
        # Regexp pattern for variables and constants to find.
        # Default: "(?i)passwd|pass|password|pwd|secret|token|pw|apiKey|bearer|cred"
        pattern: "(?i)example"
        # If true, complain about all cases (even with low entropy).
        # Default: false
        ignore_entropy: false
        # Maximum allowed entropy of the string.
        # Default: "80.0"
        entropy_threshold: "80.0"
        # Maximum allowed value of entropy/string length.
        # Is taken into account if entropy >= entropy_threshold/2.
        # Default: "3.0"
        per_char_threshold: "3.0"
        # Calculate entropy for first N chars of the string.
        # Default: "16"
        truncate: "32"
      # Additional functions to ignore while checking unhandled errors.
      # Following functions always ignored:
      #   bytes.Buffer:
      #     - Write
      #     - WriteByte
      #     - WriteRune
      #     - WriteString
      #   fmt:
      #     - Print
      #     - Printf
      #     - Println
      #     - Fprint
      #     - Fprintf
      #     - Fprintln
      #   strings.Builder:
      #     - Write
      #     - WriteByte
      #     - WriteRune
      #     - WriteString
      #   io.PipeWriter:
      #     - CloseWithError
      #   hash.Hash:
      #     - Write
      #   os:
      #     - Unsetenv
      # Default: {}
      G104:
        fmt:
          - Fscanf
      G111:
        # Regexp pattern to find potential directory traversal.
        # Default: "http\\.Dir\\(\"\\/\"\\)|http\\.Dir\\('\\/'\\)"
        pattern: "custom\\.Dir\\(\\)"
      # Maximum allowed permissions mode for os.Mkdir and os.MkdirAll
      # Default: "0750"
      G301: "0750"
      # Maximum allowed permissions mode for os.OpenFile and os.Chmod
      # Default: "0600"
      G302: "0600"
      # Maximum allowed permissions mode for os.WriteFile and ioutil.WriteFile
      # Default: "0600"
      G306: "0600"

  lll:
    # Max line length, lines longer will be reported.
    # '\t' is counted as 1 character by default, and can be changed with the tab-width option.
    # Default: 120.
    line-length: 120
    # Tab width in spaces.
    # Default: 1
    tab-width: 1

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
    - nolintlint
    - staticcheck
    - makezero
    - nilerr
    - errorlint
    - bodyclose
    - exportloopref
    - gci
    - gosec
    - lll
  fast: false
`

var makefileConfig = `
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

.PHONY: pre-commite
pre-commite: install lint tests build

.PHONY: ci
ci: pre-commite

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