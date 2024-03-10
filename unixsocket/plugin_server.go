/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unixsocket

import (
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/unixsocket/server"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type (
	Config struct {
		Path string `yaml:"unix"`
	}
)

func (v *Config) Default() {
	v.Path = "./app.socket"
}

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l xlog.Logger) Server {
			return newServerProvider(c, l)
		},
	}
}

type (
	serverProvider struct {
		config *Config
		serv   *server.Server
		wg     iosync.Group
		log    xlog.Logger
	}

	Server interface {
		Command(name string, h func([]byte) ([]byte, error))
	}
)

func newServerProvider(c *Config, l xlog.Logger) *serverProvider {
	return &serverProvider{
		config: c,
		log:    l,
		serv:   server.New(c.Path),
	}
}

func (v *serverProvider) Up(ctx xc.Context) error {
	v.serv.ErrorLog(func(err error) {
		v.log.WithError("err", err).Errorf("unix")
	})
	v.wg.Background(func() {
		if err := v.serv.Up(); err != nil {
			v.log.WithFields(xlog.Fields{
				"err":  err.Error(),
				"path": v.config.Path,
			}).Errorf("Unix server stopped")
			ctx.Close()
			return
		}
		v.log.WithFields(xlog.Fields{
			"path": v.config.Path,
		}).Infof("Unix server stopped")
	})
	return nil
}

func (v *serverProvider) Down() error {
	err := v.serv.Down()
	v.wg.Wait()
	return err
}

func (v *serverProvider) Command(name string, h func([]byte) ([]byte, error)) {
	v.serv.AddCommand(name, h)
}
