/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"fmt"
	"strings"
	"time"

	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/network/listen"
)

// ConfigGroup config to initialize HTTP service
type ConfigGroup struct {
	HTTP []Config `yaml:"http"`
}

type Config struct {
	Tag             string               `yaml:"tag"`
	Addr            string               `yaml:"addr"`
	Network         string               `yaml:"network,omitempty"`
	ReadTimeout     time.Duration        `yaml:"read_timeout,omitempty"`
	WriteTimeout    time.Duration        `yaml:"write_timeout,omitempty"`
	IdleTimeout     time.Duration        `yaml:"idle_timeout,omitempty"`
	ShutdownTimeout time.Duration        `yaml:"shutdown_timeout,omitempty"`
	Tls             []listen.Certificate `yaml:"tls,omitempty"`
}

func (v *ConfigGroup) Default() {
	if v.HTTP == nil {
		v.HTTP = append(v.HTTP, Config{
			Tag:  "main",
			Addr: "0.0.0.0:8080",
		})
	}
}

func (v *ConfigGroup) Validate() error {
	if v.HTTP == nil {
		return fmt.Errorf("http server: empty config")
	}

	hasMain := false
	for i, cfg := range v.HTTP {
		if cfg.Tag == "" {
			return fmt.Errorf("http server: tag is required (config=%d)", i)
		}
		if cfg.Tag == "main" {
			hasMain = true
		}
		if _, ok := networkType[cfg.Network]; !ok {
			if cfg.Network != "" {
				return fmt.Errorf(
					"http server: network '%s' is not supported, want %s",
					cfg.Network, strings.Join(do.Keys(networkType), ","),
				)
			}
		}
		if cfg.Tls != nil {
			for _, cert := range cfg.Tls {
				if !fs.FileExist(cert.CAFile) {
					return fmt.Errorf("http server: tls certificate "+
						"'%s' not exist (config=%d)", cert.CAFile, i)
				}
				if !fs.FileExist(cert.CertFile) {
					return fmt.Errorf("http server: tls certificate "+
						"'%s' not exist (config=%d)", cert.CertFile, i)
				}
				if !fs.FileExist(cert.KeyFile) {
					return fmt.Errorf("http server: tls certificate "+
						"'%s' not exist (config=%d)", cert.CertFile, i)
				}
				if cert.AutoGenerate && len(cert.Addresses) == 0 {
					return fmt.Errorf(
						"http server: tls autogenerate certificate "+
							"must have at least one address (config=%d)", i)
				}
			}
		}
	}

	if !hasMain {
		return fmt.Errorf("http server: no main config found")
	}

	return nil
}
