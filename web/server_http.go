/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"context"
	"net"
	"net/http"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/logx"
	"go.osspkg.com/network/address"
	"go.osspkg.com/syncing"
	"go.osspkg.com/xc"
)

type (
	Config struct {
		Tag             string        `yaml:"tag"`
		Addr            string        `yaml:"addr"`
		Network         string        `yaml:"network,omitempty"`
		ReadTimeout     time.Duration `yaml:"read_timeout,omitempty"`
		WriteTimeout    time.Duration `yaml:"write_timeout,omitempty"`
		IdleTimeout     time.Duration `yaml:"idle_timeout,omitempty"`
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout,omitempty"`
	}

	Server struct {
		conf    Config
		serv    *http.Server
		handler http.Handler
		wg      syncing.Group
		sync    syncing.Switch
	}
)

// NewServer create default http server
func NewServer(conf Config, handler http.Handler) *Server {
	srv := &Server{
		conf:    conf,
		handler: handler,
		sync:    syncing.NewSwitch(),
		wg:      syncing.NewGroup(),
	}
	srv.validate()
	return srv
}

func (v *Server) validate() {
	if v.conf.ReadTimeout == 0 {
		v.conf.ReadTimeout = defaultTimeout
	}
	if v.conf.WriteTimeout == 0 {
		v.conf.WriteTimeout = defaultTimeout
	}
	if v.conf.IdleTimeout == 0 {
		v.conf.IdleTimeout = defaultTimeout
	}
	if v.conf.ShutdownTimeout == 0 {
		v.conf.ShutdownTimeout = defaultShutdownTimeout
	}
	if len(v.conf.Network) == 0 {
		v.conf.Network = defaultNetwork
	}
	if _, ok := networkType[v.conf.Network]; !ok {
		v.conf.Network = defaultNetwork
	}
	v.conf.Addr = address.ResolveIPPort(v.conf.Addr)
}

// Up start http server
func (v *Server) Up(ctx xc.Context) error {
	if !v.sync.On() {
		return errors.Wrapf(errServAlreadyRunning, "starting server on %s", v.conf.Addr)
	}
	v.serv = &http.Server{
		ReadTimeout:  v.conf.ReadTimeout,
		WriteTimeout: v.conf.WriteTimeout,
		IdleTimeout:  v.conf.IdleTimeout,
		Handler:      v.handler,
	}

	nl, err := net.Listen(v.conf.Network, v.conf.Addr)
	if err != nil {
		return err
	}

	logx.Info("Http server started", "tag", v.conf.Tag, "ip", v.conf.Addr)

	v.wg.Background(func() {
		defer ctx.Close()
		if err = v.serv.Serve(nl); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Error("Http server stopped", "tag", v.conf.Tag, "ip", v.conf.Addr, "err", err)
			return
		}

		logx.Info("Http server stopped", "tag", v.conf.Tag, "ip", v.conf.Addr)
	})
	return nil
}

// Down stop http server
func (v *Server) Down() error {
	if !v.sync.Off() {
		return errors.Wrapf(errServAlreadyStopped, "stopping server on %s", v.conf.Addr)
	}
	ctx, cncl := context.WithTimeout(context.Background(), v.conf.ShutdownTimeout)
	defer cncl()
	err := v.serv.Shutdown(ctx)
	v.wg.Wait()
	return err
}
