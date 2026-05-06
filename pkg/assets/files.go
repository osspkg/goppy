/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package assets

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.osspkg.com/errors"
)

const FileSuffix = "_assets.go"

var (
	ErrUndefinedFormat = errors.New("undefined format")
)

func (c *Cache) fileCreate(filename string, call func(r io.Writer) error) error {
	fi, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fi.Close() //nolint:errcheck
	return call(fi)
}

func (c *Cache) fileOpen(filename string, call func(r io.Reader) error) error {
	fi, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fi.Close() //nolint:errcheck
	return call(fi)
}

func (c *Cache) FromFile(filename string) error {
	switch true {
	case strings.HasSuffix(filename, ".tar.gz"):
		return c.fileOpen(filename, func(v io.Reader) error {
			return c.FromTGZ(v)
		})

	case strings.HasSuffix(filename, ".tar"):
		return c.fileOpen(filename, func(v io.Reader) error {
			return c.FromTar(v)
		})

	default:
		return ErrUndefinedFormat
	}
}

func (c *Cache) FromDir(dir string) error {
	if v, err := filepath.Abs(dir); err == nil {
		dir = v
	} else {
		return err
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, FileSuffix) {
			return nil
		}
		if b, err := os.ReadFile(path); err != nil {
			return err
		} else {
			path = strings.TrimPrefix(path, dir)
			c.Set(path, b)
		}
		return nil
	})
}

func (c *Cache) ToFile(filename string) error {
	switch {
	case strings.HasSuffix(filename, ".tar.gz"):
		return c.fileCreate(filename, func(w io.Writer) error {
			return c.ToTGZ(w)
		})

	case strings.HasSuffix(filename, ".tar"):
		return c.fileCreate(filename, func(w io.Writer) error {
			return c.ToTar(w)
		})

	default:
		return ErrUndefinedFormat
	}
}
