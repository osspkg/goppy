/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package tcp

import (
	"time"

	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/plugins"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type Config struct {
	TCP []ConfigItem `yaml:"tcp"`
}

func (v *Config) Default() {
	if len(v.TCP) == 0 {
		v.TCP = append(v.TCP, ConfigItem{
			Address: "0.0.0.0:8080",
			Certs: []Cert{{
				Public:  "./ssl/public.crt",
				Private: "./ssl/private.key",
			}},
			Timeout: 5 * time.Second,
		})
	}
}

type (
	Server interface {
		HandleFunc(h HandlerTCP)
	}

	serverProvider struct {
		log   xlog.Logger
		conf  []ConfigItem
		servs []*ServerTCP
		wg    iosync.Group
	}
)

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &Config{},
		Inject: func(c *Config, l xlog.Logger) Server {
			return &serverProvider{
				log:   l,
				conf:  c.TCP,
				servs: make([]*ServerTCP, 0, len(c.TCP)),
				wg:    iosync.NewGroup(),
			}
		},
	}
}

func (v *serverProvider) HandleFunc(h HandlerTCP) {
	for _, serv := range v.servs {
		serv.HandleFunc(h)
	}
}

func (v *serverProvider) Up(ctx xc.Context) error {
	for _, conf := range v.conf {
		conf := conf
		serv := NewServerTCP(conf)
		v.servs = append(v.servs, serv)
		v.log.WithFields(xlog.Fields{
			"addr": conf.Address,
		}).Infof("TCP server started")
		v.wg.Background(func() {
			if err := serv.ListenAndServe(ctx.Context()); err != nil {
				v.log.WithFields(xlog.Fields{
					"err": err.Error(), "addr": conf.Address,
				}).Errorf("TCP server stopped")
				ctx.Close()
				return
			}
			v.log.WithFields(xlog.Fields{
				"addr": conf.Address,
			}).Infof("TCP server stopped")
		})
	}
	return nil
}

func (v *serverProvider) Down() error {
	var err error
	for _, serv := range v.servs {
		err = errors.Wrap(err, serv.Close())
	}
	v.wg.Wait()
	return err
}
