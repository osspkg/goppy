/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type (
	Resolver interface {
		Name() string
		Value(name string) (string, bool)
	}

	Config struct {
		blob      []byte
		resolvers []Resolver
	}
)

func NewConfigResolve(rs ...Resolver) *Config {
	return &Config{
		blob:      nil,
		resolvers: rs,
	}
}

func (v *Config) OpenFile(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	v.blob = b
	return nil
}

func (v *Config) Decode(cs ...interface{}) error {
	for _, c := range cs {
		if err := yaml.Unmarshal(v.blob, c); err != nil {
			return err
		}
	}
	return nil
}

var rexName = regexp.MustCompile(`(?m)^[a-z][a-z0-9]+$`)

func (v *Config) Build() error {
	for _, resolver := range v.resolvers {
		if !rexName.MatchString(resolver.Name()) {
			return fmt.Errorf("resolver '%s' has invalid name, must like regexp [a-z][a-z0-9]+", resolver.Name())
		}
		rex := regexp.MustCompile(fmt.Sprintf(`(?mUsi)@%s\((.+)#(.*)\)`, resolver.Name()))
		submatchs := rex.FindAllSubmatch(v.blob, -1)

		for _, submatch := range submatchs {
			pattern, key, defval := submatch[0], submatch[1], submatch[2]

			if val, ok := resolver.Value(string(key)); ok && len(val) > 0 {
				defval = []byte(val)
			}

			v.blob = bytes.ReplaceAll(v.blob, pattern, defval)
		}
	}

	return nil
}
