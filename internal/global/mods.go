/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import (
	"os"
	"regexp"
	"strings"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/validate"
)

var rexModule = regexp.MustCompile(`(?mU)module (.*)\n`)
var SkipErr = errors.New("skip read module")

type Module struct {
	Name    string
	File    string
	Prefix  string
	Version *validate.Version
	Changed bool
}

func SearchModule(dir string) (map[string]*Module, error) {
	list := make(map[string]*Module, 20)
	mods, err := fs.SearchFiles(dir, "go.mod")
	if err != nil {
		return nil, err
	}
	var b []byte
	for _, mod := range mods {
		if b, err = os.ReadFile(mod); err != nil {
			return nil, err
		}

		temp := rexModule.FindStringSubmatch(string(b))
		if len(temp) != 2 {
			continue
		}
		module := temp[1]
		if !strings.Contains(module, "/") {
			continue
		}
		list[module] = &Module{
			Name: module,
			File: mod,
		}
	}

	return list, nil
}

func ReadModule(filepath string) (*Module, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	temp := rexModule.FindStringSubmatch(string(b))
	if len(temp) != 2 {
		return nil, SkipErr
	}
	module := temp[1]
	if !strings.Contains(module, "/") {
		return nil, SkipErr
	}
	return &Module{
		Name: module,
		File: filepath,
	}, nil
}
