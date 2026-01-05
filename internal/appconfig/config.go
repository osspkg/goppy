/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package appconfig

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"

	"go.osspkg.com/config"
	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/codec"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/internal/applog"
	"go.osspkg.com/goppy/v3/plugins"
)

func Recovery(filename string, configs []any) error {
	if len(filename) == 0 || fs.FileExist(filename) {
		return nil
	}

	for _, cfg := range configs {
		if vv, ok := cfg.(plugins.Defaulter); ok {
			vv.Default()
			continue
		}

		if vv, ok := cfg.(plugins.Defaulter2); ok {
			if err := vv.Default(); err != nil {
				return err
			}
		}
	}

	cfg := &applog.GroupConfig{
		Log: applog.Config{
			Level:    4,
			FilePath: "/dev/stdout",
			Format:   "string",
		},
	}

	return codec.FileEncoder(filename).Encode(append(configs, cfg)...)
}

type Config struct {
	Filepath  string
	Data, Ext string
}

func DecodeAndValidate(c Config, resolvers []config.Resolver, configs []any) error {
	rc := config.New(resolvers...)

	switch {
	case len(c.Filepath) > 0:
		if err := rc.OpenFile(c.Filepath); err != nil {
			return err
		}
	case len(c.Data) > 0:
		rc.OpenBlob(c.Data, c.Ext)
	default:
		return nil
	}

	if err := rc.Build(); err != nil {
		return err
	}

	for _, cfg := range configs {
		if err := rc.Decode(cfg); err != nil {
			return fmt.Errorf("decode config %T error: %w", cfg, err)
		}
		vv, ok := cfg.(plugins.Validator)
		if !ok {
			continue
		}
		if err := vv.Validate(); err != nil {
			return fmt.Errorf("validate config %T error: %w", cfg, err)
		}
	}
	return nil
}

func CreatePID(filepath string) error {
	fi, err := os.Create(filepath)
	if err != nil {
		return err
	}
	pid := strconv.Itoa(syscall.Getpid())
	_, err = io.WriteString(fi, pid)
	return errors.Wrap(err, fi.Close())
}
