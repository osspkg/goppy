/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package unix

import (
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/sdk/app"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
	"go.osspkg.com/goppy/sdk/netutil/unixsocket"
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
		Inject: func(c *Config, l log.Logger) (*serverProvider, Server) {
			s := newServerProvider(c, l)
			return s, s
		},
	}
}

type (
	serverProvider struct {
		config *Config
		serv   *unixsocket.Server
		wg     iosync.Group
		log    log.Logger
	}

	Server interface {
		Command(name string, h func([]byte) ([]byte, error))
	}
)

func newServerProvider(c *Config, l log.Logger) *serverProvider {
	return &serverProvider{
		config: c,
		log:    l,
		serv:   unixsocket.NewServer(c.Path),
	}
}

func (v *serverProvider) Up(ctx app.Context) (err error) {
	v.serv.ErrorLog(func(err error) {
		v.log.WithError("err", err).Errorf("unix")
	})
	v.wg.Background(func() {
		if err := v.serv.Up(); err != nil {
			v.log.WithFields(log.Fields{
				"err":  err.Error(),
				"path": v.config.Path,
			}).Errorf("Unix server stopped")
			ctx.Close()
			return
		}
		v.log.WithFields(log.Fields{
			"path": v.config.Path,
		}).Infof("Unix server stopped")
	})
	return
}

func (v *serverProvider) Down() error {
	err := v.serv.Down()
	v.wg.Wait()
	return err
}

func (v *serverProvider) Command(name string, h func([]byte) ([]byte, error)) {
	v.serv.AddCommand(name, h)
}
