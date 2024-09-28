/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import (
	"context"
	"os"
	"regexp"

	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/ioutils/shell"
)

const (
	DirTools   = ".tools"
	DirBuild   = "build"
	DirInit    = "init"
	DirScripts = "scripts"
)

func GetToolsDir() string {
	return fs.CurrentDir() + "/" + DirTools
}

func GetBuildDir() string {
	return fs.CurrentDir() + "/" + DirBuild
}

func GetInitDir() string {
	return fs.CurrentDir() + "/" + DirInit
}

func GetScriptsDir() string {
	return fs.CurrentDir() + "/" + DirScripts
}

func SetupEnv() {
	console.FatalIfErr(os.Setenv("GOBIN", GetToolsDir()), "setup env")
	console.FatalIfErr(os.Setenv("PATH", GetToolsDir()+":"+os.Getenv("PATH")), "setup env")
}

var rex = regexp.MustCompile(`go(\d+)\.(\d+)`)

func GoVersion() string {
	sh := shell.New()
	sh.SetDir(fs.CurrentDir())
	console.FatalIfErr(sh.SetShell("sh", "c"), "init shell")
	b, err := sh.Call(context.TODO(), "go version")
	console.FatalIfErr(err, "detect go version")
	result := rex.FindAllString(string(b), 1)
	for _, s := range result {
		return s
	}
	return "unknown"
}
