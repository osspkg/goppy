/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.osspkg.com/algorithms/graph/kahn"
	"go.osspkg.com/console"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/validate"
	"golang.org/x/mod/modfile"

	"go.osspkg.com/goppy/v2/internal/global"
)

func Cmd() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("tag", "Create release tags")
		setter.Flag(func(fs console.FlagsSetter) {
			fs.Bool("minor", "update minor version (default - patch)")
		})
		setter.ExecFunc(func(_ []string, minor bool) {
			console.Infof("--- READ CONFIG ---")

			var (
				allMods map[string]*global.Module
				currmod *global.Module
				err     error
				b       []byte
				f       *modfile.File
				fi      os.FileInfo
				HEAD    string
			)

			console.Infof("--- GET ALL MODULES ---")

			allMods, err = global.SearchModule(fs.CurrentDir())
			console.FatalIfErr(err, "Detect go.mod files")

			var root *global.Module
			for _, m := range allMods {
				if root == nil {
					root = m
					continue
				}
				if len(root.Name) > len(m.Name) {
					root = m
				}
			}
			for _, m := range allMods {
				m.Prefix = strings.Trim(strings.TrimPrefix(m.Name, root.Name), "/")
				if len(m.Prefix) > 0 {
					m.Prefix += "/"
				}
				b, err = global.Exec("git tag -l \"" + m.Prefix + "v*\"")
				console.FatalIfErr(err, "Get tags for: %s", m.Name)
				m.Version = validate.MaxVersion(strings.Split(string(b), "\n")...)
			}

			console.Infof("--- DETECT CHANGES ---")

			HEAD, err = global.GitHEAD("")
			console.FatalIfErr(err, "Get git HEAD")
			b, err = global.Exec("git diff --name-only --ignore-submodules" +
				" --diff-filter=ACMRTUXB origin/" +
				HEAD + " -- \"*.go\" \"*.mod\" \"*.sum\"")
			console.FatalIfErr(err, "Detect changed files")
			changedFiles := strings.Split(string(b), "\n")
			for _, file := range changedFiles {
				if len(file) == 0 {
					continue
				}
				dir := filepath.Dir(file)
				isRoot := dir == "."
				for {
					currmod, err = global.ReadModule(dir + "/go.mod")
					if err != nil && !isRoot && strings.Contains(err.Error(), "no such file") {
						dir = filepath.Dir(dir)
						if dir != "." {
							continue
						}
						break
					}
					break
				}
				if err != nil {
					continue
				}

				for _, m := range allMods {
					if m.Name == currmod.Name && !m.Changed {
						m.Changed = true
						if minor {
							m.Version.Minor++
							m.Version.Patch = 0
						} else {
							m.Version.Patch++
						}
					}
				}
			}

			console.Infof("--- UPDATE MODULES ---")

			graph := kahn.New()
			for _, m := range allMods {
				_, err = os.Stat(m.File)
				console.FatalIfErr(err, "Get info go.mod file: %s", m.File)
				b, err = os.ReadFile(m.File)
				console.FatalIfErr(err, "Read go.mod file: %s", m.File)
				_, err = modfile.Parse(m.File, b, func(path, version string) (string, error) {
					if _, ok := allMods[path]; ok {
						graph.Add(path, m.Name)
					}
					return version, nil
				})
				console.FatalIfErr(err, "Parse go.mod file: %s", m.File)
			}
			console.FatalIfErr(graph.Build(), "Build graph")
			for _, s := range graph.Result() {
				m, ok := allMods[s]
				if !ok {
					continue
				}
				fmt.Println(">", m.Name)
				fi, err = os.Stat(m.File)
				console.FatalIfErr(err, "Get info go.mod file: %s", m.File)
				b, err = os.ReadFile(m.File)
				console.FatalIfErr(err, "Read go.mod file: %s", m.File)
				f, err = modfile.Parse(m.File, b, func(path, version string) (string, error) {
					if mm, ok := allMods[path]; ok && mm.Version.String() != version {
						if !m.Changed {
							if minor {
								m.Version.Minor++
								m.Version.Patch = 0
							} else {
								m.Version.Patch++
							}
							m.Changed = true
						}
						return mm.Version.String(), nil
					}
					return version, nil
				})
				console.FatalIfErr(err, "Parse go.mod file: %s", m.File)
				b, err = f.Format()
				console.FatalIfErr(err, "Format go.mod file: %s", m.File)
				err = os.WriteFile(m.File, b, fi.Mode())
				console.FatalIfErr(err, "Update go.mod file: %s", m.File)
			}

			console.Infof("--- GIT COMMITTED ---")

			cmds := make([]string, 0, 50)
			cmds = append(cmds,
				"git add .",
				"git commit -m \"release new versions\"",
			)
			for _, m := range allMods {
				if !m.Changed {
					continue
				}
				cmds = append(cmds, "git tag "+m.Prefix+m.Version.String())
			}
			cmds = append(cmds, "git push", "git push --tags")
			global.ExecPack(cmds...)
		})
	})
}
