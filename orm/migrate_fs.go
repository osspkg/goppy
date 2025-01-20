/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.osspkg.com/ioutils/fs"
)

type MFS interface {
	Done()
	Next() bool
	Dialect() string
	Tags() []string
	FileNames() ([]string, error)
	FileData(filename string) (string, error)
}

// ---------------------------------------------------------------------------------------------------------------------

type (
	memFS struct {
		conf []Migration
		curr int
	}
	Migration struct {
		Tags    []string
		Dialect string
		Data    map[string]string
	}
)

func newMemFS(c []Migration) *memFS {
	return &memFS{
		conf: c,
		curr: -1,
	}
}

func (o *memFS) Done() {
	o.curr = -1
}

func (o *memFS) Dialect() string {
	return o.conf[o.curr].Dialect
}

func (o *memFS) Next() bool {
	if len(o.conf) <= 0 {
		return false
	}
	o.curr++
	if len(o.conf) <= o.curr {
		return false
	}
	return true
}

func (o *memFS) Tags() []string {
	return o.conf[o.curr].Tags
}

func (o *memFS) FileNames() ([]string, error) {
	list := make([]string, 0)
	for name := range o.conf[o.curr].Data {
		list = append(list, name)
	}
	sort.Strings(list)
	return list, nil
}

func (o *memFS) FileData(filename string) (string, error) {
	b, ok := o.conf[o.curr].Data[filename]
	if !ok {
		return "", fmt.Errorf("not found: %s", filename)
	}
	return b, nil
}

// ---------------------------------------------------------------------------------------------------------------------

type osFS struct {
	conf []ConfigMigrateItem
	curr int
}

func newOSFS(c []ConfigMigrateItem) *osFS {
	return &osFS{
		conf: c,
		curr: -1,
	}
}

func (o *osFS) Done() {
	o.curr = -1
}

func (o *osFS) Dialect() string {
	return o.conf[o.curr].Dialect
}

func (o *osFS) Next() bool {
	if len(o.conf) <= 0 {
		return false
	}
	for {
		o.curr++
		if len(o.conf) <= o.curr {
			return false
		}
		if !fs.FileExist(o.conf[o.curr].Dir) {
			continue
		}
		return true
	}
}

func (o *osFS) Tags() []string {
	return strings.Split(o.conf[o.curr].Tags, ",")
}

func (o *osFS) FileNames() ([]string, error) {
	list, err := filepath.Glob(o.conf[o.curr].Dir + "/*.sql")
	if err != nil {
		return nil, err
	}
	sort.Strings(list)
	return list, nil
}

func (o *osFS) FileData(filename string) (string, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
