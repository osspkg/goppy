/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"go.osspkg.com/ioutils/codec"
)

type (
	Resolver interface {
		Name() string
		Resolve(keys ...string) (map[string][]byte, error)
	}

	Config struct {
		data *codec.BlobEncoder
		list []Resolver
	}

	source struct {
		Key     string
		Pattern []byte
		Default []byte
	}
)

func New(list ...Resolver) *Config {
	return &Config{
		data: &codec.BlobEncoder{
			Blob: make([]byte, 0),
			Ext:  "",
		},
		list: list,
	}
}

func (v *Config) Flush() {
	v.data.Blob = make([]byte, 0)
	v.data.Ext = ""
}

func (v *Config) OpenBlob(b, ext string) {
	v.data.Blob = []byte(b)
	v.data.Ext = ext
}

func (v *Config) OpenFile(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	v.data.Blob = b
	v.data.Ext = filepath.Ext(filename)
	return nil
}

func (v *Config) Decode(configs ...interface{}) error {
	if len(v.data.Blob) == 0 || len(v.data.Ext) == 0 {
		return fmt.Errorf("config is empty")
	}

	return v.data.Decode(configs...)
}

var rexName = regexp.MustCompile(`(?m)^[a-z][a-z0-9]+$`)

func (v *Config) Build() error {
	if len(v.data.Blob) == 0 || len(v.data.Ext) == 0 {
		return fmt.Errorf("config is empty")
	}

	for _, r := range v.list {
		if !rexName.MatchString(r.Name()) {
			return fmt.Errorf("resolver '%s' has invalid name, must like regexp [a-z][a-z0-9]+", r.Name())
		}
		rex := regexp.MustCompile(fmt.Sprintf(`(?mUsi)@%s\((.+)#(.*)\)`, r.Name()))
		submatchs := rex.FindAllSubmatch(v.data.Blob, -1)

		sources := make([]source, 0, len(submatchs))
		keys := make([]string, 0, len(submatchs))

		for _, submatch := range submatchs {
			sources = append(sources, source{
				Key:     string(submatch[1]),
				Pattern: submatch[0],
				Default: submatch[2],
			})
			keys = append(keys, string(submatch[1]))
		}

		values, err := r.Resolve(keys...)
		if err != nil {
			return fmt.Errorf("resolver '%s': %w", r.Name(), err)
		}

		for _, s := range sources {
			val := bytes.TrimSpace(values[s.Key])
			if len(val) < 1 {
				val = s.Default
			}

			v.data.Blob = bytes.ReplaceAll(v.data.Blob, s.Pattern, val)
		}
	}
	return nil
}
