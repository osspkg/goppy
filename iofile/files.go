/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package iofile

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func Search(dir, filename string) ([]string, error) {
	files := make([]string, 0, 2)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != filename {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func SearchByExt(dir, ext string) ([]string, error) {
	files := make([]string, 0, 2)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(info.Name()) != ext {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func Rewrite(filename string, call func([]byte) ([]byte, error)) error {
	if !Exist(filename) {
		if err := os.WriteFile(filename, []byte(""), 0755); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if b, err = call(b); err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0755)
}

func Copy(dst, src string, mode os.FileMode) error {
	source, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer source.Close() //nolint: errcheck

	if mode == 0 {
		fi, err0 := source.Stat()
		if err0 != nil {
			return err0
		}
		mode = fi.Mode()
	}

	dist, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dist.Close() //nolint: errcheck

	_, err = io.Copy(dist, source)
	return err
}

func Folder(filename string) string {
	dir := filepath.Dir(filename)
	tree := strings.Split(dir, string(os.PathSeparator))
	return tree[len(tree)-1]
}
