/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package formdata

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"go.osspkg.com/errors"
)

type mpReader interface {
	ParseMultipartForm(int64) error
	FormValue(key string, resolve func(value string) error) error
	FormFile(key string, resolve func(value io.Reader, size int64) error) error
}

type multipartReadAdapter struct {
	r *multipart.Reader
	f *multipart.Form
}

func newMultipartReadAdapter(r io.Reader, boundary string) *multipartReadAdapter {
	return &multipartReadAdapter{
		r: multipart.NewReader(r, boundary),
	}
}

func (a *multipartReadAdapter) Close() error {
	if a.f == nil {
		return nil
	}

	return a.f.RemoveAll()
}

func (a *multipartReadAdapter) ParseMultipartForm(maxMemory int64) error {
	var err error
	a.f, err = a.r.ReadForm(maxMemory)
	return err
}

func (a *multipartReadAdapter) FormValue(key string, resolve func(value string) error) error {
	if a.f == nil {
		if err := a.ParseMultipartForm(defaultMaxMemory); err != nil {
			return err
		}
	}
	if v, ok := a.f.Value[key]; ok && len(v) > 0 {
		return resolve(v[0])
	}
	return fmt.Errorf("field `%s` not found", key)
}

func (a *multipartReadAdapter) FormFile(key string, resolve func(value io.Reader, size int64) error) error {
	if a.f == nil {
		if err := a.ParseMultipartForm(defaultMaxMemory); err != nil {
			return err
		}
	}
	if f, ok := a.f.File[key]; ok && len(f) > 0 {
		fi, err := f[0].Open()
		if err != nil {
			return err
		}
		err = resolve(fi, f[0].Size)
		return errors.Wrap(err, fi.Close())
	}
	return errMissingFile
}

func getBoundary(r io.Reader) (string, error) {
	boundary, err := bufio.NewReaderSize(r, 512).ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(boundary) < 2 {
		return "", errors.New("boundary too short")
	}
	if boundary[:2] != "--" {
		return "", errors.New("boundary does not start with --")
	}
	return strings.TrimSpace(boundary[2:]), nil
}
